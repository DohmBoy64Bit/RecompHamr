[CmdletBinding()]
param(
    [string]$SkillsRoot = (Join-Path $PSScriptRoot '../../internal/skills/builtin'),
    [string]$ArtifactDirectory = 'E:\ReProject\StageG-Evaluation',
    [string]$BaseUrl = 'http://localhost:1234',
    [string]$Model = 'mistralai/devstral-small-2-2512',
    [ValidateRange(1, 10)][int]$Trials = 3,
    [Alias('Skill')][string[]]$SkillName,
    [ValidateRange(30, 1800)][int]$TimeoutSeconds = 300,
    [switch]$ValidateOnly,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Get-RequiredProperty($Object, [string]$Name, [string]$Context) {
    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property -or $null -eq $property.Value) { throw "$Context requires '$Name'" }
    return $property.Value
}

function Read-Skill([IO.FileInfo]$File) {
    $text = [IO.File]::ReadAllText($File.FullName)
    $normalized = $text.Replace("`r`n", "`n")
    if (-not $normalized.StartsWith("---`n")) { throw "$($File.FullName) has no YAML frontmatter" }
    $end = $normalized.IndexOf("`n---`n", 4, [StringComparison]::Ordinal)
    if ($end -lt 0) { throw "$($File.FullName) has unterminated YAML frontmatter" }
    $frontmatter = $normalized.Substring(4, $end - 4)
    $values = @{}
    foreach ($line in $frontmatter -split "`n") {
        if ($line -match '^([a-z-]+):\s*(.*)$') { $values[$Matches[1]] = $Matches[2].Trim().Trim('"').Trim("'") }
    }
    foreach ($required in @('name', 'description')) {
        if (-not $values.ContainsKey($required) -or [string]::IsNullOrWhiteSpace($values[$required])) { throw "$($File.FullName) requires '$required'" }
    }
    if ($values.name -cne $File.Directory.Name) { throw "$($File.FullName) name does not match its directory" }
    $evalPath = Join-Path $File.Directory.FullName 'evals/evals.json'
    if (-not (Test-Path -LiteralPath $evalPath -PathType Leaf)) { throw "$($values.name) has no evals/evals.json" }
    $eval = Get-Content -LiteralPath $evalPath -Raw | ConvertFrom-Json
    if ((Get-RequiredProperty $eval 'skill' $evalPath) -cne $values.name) { throw "$evalPath names a different skill" }
    $triggers = @(Get-RequiredProperty $eval 'trigger_cases' $evalPath)
    $outputs = @(Get-RequiredProperty $eval 'output_cases' $evalPath)
    if ($triggers.Count -lt 16 -or @($triggers | Where-Object should_trigger).Count -lt 8 -or @($triggers | Where-Object { -not $_.should_trigger }).Count -lt 8) {
        throw "$evalPath requires at least eight positive and eight negative trigger cases"
    }
    if ($outputs.Count -lt 3) { throw "$evalPath requires at least three output cases" }
    foreach ($case in $triggers) {
        [void](Get-RequiredProperty $case 'prompt' "$evalPath trigger case")
        [void](Get-RequiredProperty $case 'should_trigger' "$evalPath trigger case")
    }
    foreach ($case in $outputs) {
        [void](Get-RequiredProperty $case 'prompt' "$evalPath output case")
        [void](Get-RequiredProperty $case 'expected_outcome' "$evalPath output case")
    }
    $instructions = $normalized.Substring($end + 5).Trim()
    [pscustomobject]@{
        Name = [string]$values.name
        Description = [string]$values.description
        Instructions = $instructions
        TriggerCases = $triggers
        OutputCases = $outputs
        TriggerIdentity = ($triggers | ConvertTo-Json -Depth 8 -Compress)
        OutputIdentity = ($instructions + ':' + ($outputs | ConvertTo-Json -Depth 8 -Compress))
        LegacyIdentity = ((Get-FileHash -LiteralPath $File.FullName -Algorithm SHA256).Hash + ':' + (Get-FileHash -LiteralPath $evalPath -Algorithm SHA256).Hash)
    }
}

function Invoke-Model(
    [string]$System,
    [string]$User,
    [int]$MaxTokens = 2048,
    [hashtable]$JsonSchema,
    [ValidateSet('none', 'low', 'medium', 'high')][string]$ReasoningEffort = 'low'
) {
    $headers = @{}
    if (-not [string]::IsNullOrWhiteSpace($env:RECOMPHAMR_EVAL_API_KEY)) { $headers.Authorization = "Bearer $($env:RECOMPHAMR_EVAL_API_KEY)" }
    $request = @{
        model = $Model; stream = $false; temperature = 0; seed = 424242; max_tokens = $MaxTokens; reasoning_effort = $ReasoningEffort
        messages = @(@{ role = 'system'; content = $System }, @{ role = 'user'; content = $User })
    }
    if ($null -ne $JsonSchema) { $request.response_format = @{ type = 'json_schema'; json_schema = @{ name = 'evaluation_result'; strict = $true; schema = $JsonSchema } } }
    $body = $request | ConvertTo-Json -Depth 12 -Compress
    $response = Invoke-RestMethod -Method Post -Uri ($BaseUrl.TrimEnd('/') + '/v1/chat/completions') -Headers $headers -ContentType 'application/json' -Body $body -TimeoutSec $TimeoutSeconds
    $message = $response.choices[0].message
    $content = [string]$message.content
    if ([string]::IsNullOrWhiteSpace($content) -and $ReasoningEffort -ne 'none') {
        return Invoke-Model $System $User $MaxTokens -JsonSchema $JsonSchema -ReasoningEffort none
    }
    if ([string]::IsNullOrWhiteSpace($content)) { throw 'model returned an empty response' }
    return $content
}

function ConvertFrom-ModelJson([string]$Text, [string]$Context) {
    $candidate = $Text.Trim()
    $candidates = New-Object Collections.Generic.List[string]
    $candidates.Add($candidate)
    foreach ($match in [regex]::Matches($candidate, '(?s)```(?:json)?\s*(.*?)\s*```')) { $candidates.Add($match.Groups[1].Value) }
    $arrayStart, $arrayEnd = $candidate.IndexOf('['), $candidate.LastIndexOf(']')
    if ($arrayStart -ge 0 -and $arrayEnd -gt $arrayStart) { $candidates.Add($candidate.Substring($arrayStart, $arrayEnd - $arrayStart + 1)) }
    $objects = [regex]::Matches($candidate, '(?s)\{[^{}]*\}')
    for ($i = $objects.Count - 1; $i -ge 0; $i--) { $candidates.Add($objects[$i].Value) }
    for ($i = 0; $i -lt $candidates.Count; $i++) {
        try { return $candidates[$i] | ConvertFrom-Json } catch {}
    }
    throw "$Context returned malformed JSON"
}

$resolvedSkills = (Resolve-Path -LiteralPath $SkillsRoot).Path
$allSkills = @(Get-ChildItem -LiteralPath $resolvedSkills -Directory | Sort-Object Name | ForEach-Object { Read-Skill (Get-Item -LiteralPath (Join-Path $_.FullName 'SKILL.md')) })
if ($allSkills.Count -eq 0) { throw 'no bundled skills were found' }
$skills = $allSkills
if (@($SkillName | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }).Count -gt 0) {
    $requested = @($SkillName | ForEach-Object { $_ -split ',' } | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | ForEach-Object Trim | Sort-Object -Unique)
    $skills = @($skills | Where-Object { $_.Name -in $requested })
    $selectedNames = @($skills | ForEach-Object Name)
    $missing = @($requested | Where-Object { $_ -notin $selectedNames })
    if ($missing.Count -gt 0) { throw "unknown skill selection: $($missing -join ', ')" }
}
if ($ValidateOnly) { Write-Host "Stage G skill evaluation fixtures: PASS ($($skills.Count) skills)"; exit 0 }

$modelList = Invoke-RestMethod -Uri ($BaseUrl.TrimEnd('/') + '/v1/models') -TimeoutSec 10
if ($Model -notin @($modelList.data.id)) { throw "model '$Model' is not exposed by $BaseUrl" }
[void](New-Item -ItemType Directory -Force -Path $ArtifactDirectory)
$rawDirectory = Join-Path $ArtifactDirectory 'raw'
[void](New-Item -ItemType Directory -Force -Path $rawDirectory)
$catalog = ($allSkills | ForEach-Object { "- $($_.Name): $($_.Description)" }) -join "`n"
$catalogIdentity = ($allSkills | ForEach-Object { "$($_.Name):$($_.Description)" }) -join '|'
$started = [DateTime]::UtcNow
$summaries = New-Object Collections.Generic.List[object]
$triggerSchema = @{
    type = 'object'; additionalProperties = $false; required = @('decisions')
    properties = @{ decisions = @{ type = 'array'; items = @{ type = 'object'; additionalProperties = $false; required = @('id', 'selected'); properties = @{ id = @{ type = 'integer' }; selected = @{ type = @('string', 'null') } } } } }
}
$judgeSchema = @{
    type = 'object'; additionalProperties = $false; required = @('a_satisfies', 'b_satisfies', 'winner', 'reason_category')
    properties = @{ a_satisfies = @{ type = 'boolean' }; b_satisfies = @{ type = 'boolean' }; winner = @{ type = 'string'; enum = @('a', 'b', 'tie') }; reason_category = @{ type = 'string' } }
}

foreach ($current in $skills) {
    $safeName = $current.Name
    $rawPath = Join-Path $rawDirectory "$safeName.json"
    $candidate = $null
    $reuseTriggers = $false
    $reuseOutputs = $false
    if ((Test-Path -LiteralPath $rawPath -PathType Leaf) -and -not $Force) {
        $candidate = Get-Content -LiteralPath $rawPath -Raw | ConvertFrom-Json
        $sameModel = $candidate.model -ceq $Model
        $legacyIdentityMatches = $null -ne $candidate.PSObject.Properties['skill_identity'] -and $candidate.skill_identity -ceq $current.LegacyIdentity
        $triggerIdentityMatches = ($null -ne $candidate.PSObject.Properties['trigger_identity'] -and $candidate.trigger_identity -ceq $current.TriggerIdentity) -or $legacyIdentityMatches
        $outputIdentityMatches = ($null -ne $candidate.PSObject.Properties['output_identity'] -and $candidate.output_identity -ceq $current.OutputIdentity) -or $legacyIdentityMatches
        $reuseTriggers = $sameModel -and [int]$candidate.trials -eq $Trials -and $null -ne $candidate.PSObject.Properties['catalog_identity'] -and $candidate.catalog_identity -ceq $catalogIdentity -and $triggerIdentityMatches
        $reuseOutputs = $sameModel -and $outputIdentityMatches
    }
    if ($reuseTriggers) {
        $triggerRuns = @($candidate.trigger_runs)
    } else {
        $triggerRuns = New-Object Collections.Generic.List[object]
        $indexed = for ($i = 0; $i -lt $current.TriggerCases.Count; $i++) { [pscustomobject]@{ id = $i; prompt = [string]$current.TriggerCases[$i].prompt } }
        for ($trial = 1; $trial -le $Trials; $trial++) {
            $user = 'Choose the single matching skill for each prompt. Use null when none match. Return only a JSON object {"decisions":[{"id":0,"selected":null}]}. Valid skill names follow.' + "`n`n$catalog`n`nPROMPTS:`n" + ($indexed | ConvertTo-Json -Depth 4)
            $answer = Invoke-Model 'You are a strict skill router. Select by the documented boundary, not by superficial keywords.' $user 2048 -JsonSchema $triggerSchema -ReasoningEffort none
            $parsedObject = ConvertFrom-ModelJson $answer "$safeName trigger trial $trial"
            $parsed = @(Get-RequiredProperty $parsedObject 'decisions' "$safeName trigger trial $trial")
            if ($parsed.Count -ne $current.TriggerCases.Count) { throw "$safeName trigger trial $trial returned $($parsed.Count) decisions" }
            $triggerRuns.Add([pscustomobject]@{ trial = $trial; response = $answer; decisions = $parsed })
        }
        $triggerRuns = $triggerRuns.ToArray()
    }
    if ($reuseOutputs) {
        $outputRuns = @($candidate.output_runs)
    } else {
        $outputRuns = New-Object Collections.Generic.List[object]
        for ($i = 0; $i -lt $current.OutputCases.Count; $i++) {
            $case = $current.OutputCases[$i]
            $baseline = Invoke-Model 'Answer accurately and safely. Distinguish evidence from assumptions. Be concise.' ([string]$case.prompt) 4096
            $withSkill = Invoke-Model "Follow this task-specific skill exactly when relevant. Be concise.`n`n$($current.Instructions)" ([string]$case.prompt) 4096
            $judgeRequest = @{ expected = [string]$case.expected_outcome; answer_a = $baseline; answer_b = $withSkill } | ConvertTo-Json -Depth 6
            $judgeText = Invoke-Model 'Blindly grade two answers against the expected outcome. Return only JSON: {"a_satisfies":boolean,"b_satisfies":boolean,"winner":"a"|"b"|"tie","reason_category":"short non-sensitive label"}. Prefer a tie when quality is materially equal.' $judgeRequest 2048 -JsonSchema $judgeSchema -ReasoningEffort none
            $judge = ConvertFrom-ModelJson $judgeText "$safeName output case $i judge"
            $outputRuns.Add([pscustomobject]@{ id = $i; prompt = [string]$case.prompt; expected = [string]$case.expected_outcome; baseline = $baseline; with_skill = $withSkill; judge_response = $judgeText; judge = $judge })
        }
        $outputRuns = $outputRuns.ToArray()
    }
    $raw = [pscustomobject]@{ schema_version = 2; skill = $safeName; model = $Model; trials = $Trials; catalog_identity = $catalogIdentity; trigger_identity = $current.TriggerIdentity; output_identity = $current.OutputIdentity; trigger_runs = $triggerRuns; output_runs = $outputRuns }
    $raw | ConvertTo-Json -Depth 14 | Set-Content -LiteralPath $rawPath -Encoding utf8

    $correct = 0; $total = 0
    foreach ($run in $raw.trigger_runs) {
        foreach ($decision in $run.decisions) {
            $id = [int]$decision.id
            if ($id -lt 0 -or $id -ge $current.TriggerCases.Count) { throw "$safeName returned out-of-range trigger id $id" }
            $selected = if ($null -eq $decision.selected -or [string]::IsNullOrWhiteSpace([string]$decision.selected)) { $null } else { [string]$decision.selected }
            $correctDecision = if ([bool]$current.TriggerCases[$id].should_trigger) { $selected -ceq $safeName } else { $selected -cne $safeName }
            if ($correctDecision) { $correct++ }; $total++
        }
    }
    $skillSatisfies = @($raw.output_runs | Where-Object { [bool]$_.judge.b_satisfies }).Count
    $baselineSatisfies = @($raw.output_runs | Where-Object { [bool]$_.judge.a_satisfies }).Count
    $skillWins = @($raw.output_runs | Where-Object { $_.judge.winner -eq 'b' }).Count
    $baselineWins = @($raw.output_runs | Where-Object { $_.judge.winner -eq 'a' }).Count
    $summaries.Add([pscustomobject]@{ skill = $safeName; trigger_correct = $correct; trigger_total = $total; trigger_accuracy = [Math]::Round($correct / [double]$total, 4); output_cases = @($raw.output_runs).Count; with_skill_satisfies = $skillSatisfies; baseline_satisfies = $baselineSatisfies; with_skill_wins = $skillWins; baseline_wins = $baselineWins; human_review = 'required' })
    Write-Host "${safeName}: trigger $correct/$total; output satisfies $skillSatisfies/$(@($raw.output_runs).Count); wins $skillWins-$baselineWins"
}

$report = [ordered]@{
    schema_version = 1; status = 'model-evaluation-complete-human-review-required'; model = $Model
    base_url_category = 'loopback-openai-compatible'; started_utc = $started.ToString('o'); completed_utc = [DateTime]::UtcNow.ToString('o')
    trials = $Trials; skills = $summaries.ToArray(); raw_transcripts = 'retained privately in the sibling raw directory; not suitable for commit'
}
$report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $ArtifactDirectory 'report.json') -Encoding utf8
Write-Host "Stage G skill evaluation model runs complete; human transcript review remains required: $ArtifactDirectory"
