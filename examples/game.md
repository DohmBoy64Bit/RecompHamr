Create a single file `game.html` (HTML+CSS+JS+Three.js via CDN, NO build tools, NO external assets, NO postprocessing). On load the game world renders and simulates immediately. A single click anywhere acquires pointer lock for mouse-look – this is the only gate, no menu screen.

GAME: Minimal Quake3-style arena shooter, first-person.

============================================================
SETUP — exact technical requirements
============================================================

Three.js delivery:
- Importmap in the document (head or body, anywhere before the module script) mapping `"three"` to `https://unpkg.com/three@0.160.0/build/three.module.js`
- Import via `import * as THREE from 'three'` inside ONE `<script type="module">` block
- No EffectComposer, no postprocessing passes, no shader passes
- NO CSS vignette, NO radial darkening overlay, NO color-grade filters — canvas renders edge-to-edge cleanly

DOM / CSS skeleton:
- `<canvas id="c">` filling the viewport: `position: fixed; inset: 0; width: 100vw; height: 100vh; display: block;`
- `html, body { cursor: none; overflow: hidden; margin: 0; background: #1A2230; }` — body bg matches sky so no flash during load
- `<div id="hud">` overlay covering the viewport with `pointer-events: none` — clicks always reach the canvas
- `<div id="err">` for error display (see P2); hidden by default
- See HUD section for child elements

============================================================
TUNED CONSTANTS — single source of truth. Use these exact values.
============================================================

Sensitivity (mouse → radians per pixel):   SENS = 0.00066
                                           // tuned; 0.0022 feels like ripping your head off

Movement:
  GRAVITY        = 28      // u/s²  (firm; 9.8 feels moony)
  WALK           = 7       // u/s
  SPRINT         = 11      // u/s
  GROUND_ACC     = 90      // u/s²  (max speed in <0.1 s)
  GROUND_FRIC    = 12      // u/s²  decel when no input
  AIR_ACC        = 18      // u/s²  limited air control
  JUMP_V         = 9.0     // u/s
  COYOTE         = 0.090   // s     jump grace after leaving ground
  JUMP_BUF       = 0.120   // s     pre-land jump buffer

Player body (pos.y = EYE height):
  PLAYER_HW      = 0.35    // half-width XZ
  PLAYER_H       = 1.7     // total height
  EYE_H          = 1.6     // eye offset from feet

Bot body (pos.y = FEET height):
  BOT_HW_X       = 0.40
  BOT_HW_Z       = 0.30
  BOT_H          = 1.9
  BOT_SPEED      = 3.2     // wander speed
  BOT_FIRE_COOL  = 3.8     // s, base
  BOT_FIRE_JIT   = 0.8     // s, random jitter on top
  BOT_LOS_TIME   = 1.0     // s of continuous LOS required before firing
  BOT_SPREAD     = 0.28    // rad cone (wide on purpose — forgiving)
  BOT_RANGE      = 22      // detection / fire range
  BOT_RESPAWN    = 3.0     // s
  BOT_RESPAWN_MIN_DIST = 15 // u from player at respawn

Weapon (railgun):
  RAIL_COOL      = 0.4     // s    tuned for snappy duels — do NOT raise to 1.2

Arena:
  ARENA          = 50      // floor side length, in u
  WALL_H         = 4
  COVER_COUNT    = 8..14   // randomized, none within 4 u of spawn
  RAMPS          = 2..3
  SEED           = 1337    // deterministic placement (mulberry32)

Frame loop:
  DT_CLAMP       = [0, 0.05]  // hard upper bound prevents tunneling on tab-switch

============================================================
CRITICAL FEEL REQUIREMENTS — these MUST NOT regress
============================================================

MOUSE LOOK — perfectly smooth, never jumpy, never inverted:
- Maintain two scalars on the player: `yaw` and `pitch` (radians).
- Attach the `mousemove` listener to `document` (NOT to `canvas`) so it keeps firing
  even when the cursor leaves the canvas area before pointer-lock fully engages.
- On `mousemove` while `document.pointerLockElement === canvas`:
      yaw   -= dx * SENS
      pitch -= dy * SENS
  where `dx`/`dy` come from `e.movementX`/`e.movementY` after the safeguards in P16.
  Both signs are NEGATIVE. This gives the standard FPS convention:
  mouse right → look right, mouse up → look up. Do not flip them.
- Clamp pitch to [-Math.PI/2 + 0.001, Math.PI/2 - 0.001].
- Each frame, REBUILD the camera rotation from these scalars. Do NOT rotate the camera
  incrementally — incremental rotation accumulates float drift and causes visible jumps
  and roll:
      camera.rotation.order = 'YXZ'
      camera.rotation.y = yaw
      camera.rotation.x = pitch
      camera.rotation.z = 0
- Strict 1:1 mapping. NO smoothing, NO lerp, NO acceleration curve, NO movement buffer.
- Recoil and hit-feedback NEVER write to yaw/pitch. Camera-punch effects use a separate
  transient position/FOV offset added AFTER rotation is set — the player's aim is sacred.
- All anti-jitter rules (drop-first-event, per-event clamp, unadjustedMovement,
  single-handler) are non-optional. See P16.

POINTER LOCK:
- Request it from a user-gesture click handler only; guard against duplicate requests in
  the same frame (frameId counter — see P7).
- Always request with `{ unadjustedMovement: true }` to bypass OS-level mouse acceleration
  (it is the #1 cause of erratic deltas). Fall back to the no-arg form if the returned
  promise rejects:
      const p = canvas.requestPointerLock({ unadjustedMovement: true });
      if (p && p.catch) p.catch(() => canvas.requestPointerLock());
- On `pointerlockchange` with null target → auto-pause; on re-lock → resume AND arm a
  "drop-next-mousemove" flag (see P16).
- Do NOT bind ESC as a separate pause key — the browser already releases lock on ESC,
  and a second handler would double-toggle (see P7).

MOVEMENT — snappy and grounded, not floaty, not wobbly:
- Build the horizontal input vector from WASD in camera-YAW space ONLY (ignore pitch).
  Exact formula in P13 — get the signs right.
- Apply acceleration toward `desired = wish * maxSpeed` via clamped delta:
      delta = desired - currentHV;  if (|delta| > acc*dt) delta.setLength(acc*dt);  hv += delta
  Do NOT lerp/spring-smooth `hv` toward `desired` — that creates the schwabbelig feel.
- Ground friction applies when grounded AND no input.
- Gravity integrates straight; no damping.
- If any NaN slips into `player.vel` or `player.pos`, zero/reset that tick (see P9).

JUMP & LANDING — crisp, with subtle landing dip but zero input lag:
- Jump on Space when (grounded) OR (left ground within COYOTE seconds ago)
- Buffered: if Space was pressed within JUMP_BUF seconds before touchdown, jump on landing
- Landing dip: when vertical velocity at touchdown is < -6 u/s, set `cameraOffsetY = -0.12`,
  spring back to 0 with critical damping over ~180 ms. Pure visual; never affects controls.

COLLISION — player must NEVER get stuck inside a wall, box, or corner:
- Each frame for each body (player AND bots), in this order:
    0. DEPENETRATION PASS — if AABB already overlaps any static AABB (spawn glitch,
       ramp edge, lag spike), push out along the smallest-penetration axis.
    1. Add velocity.x*dt to position.x; for each overlap, push out along X by min
       penetration and zero velocity.x.
    2. Same for Z.
    3. Same for Y; +Y stop = ceiling, -Y stop = grounded (latch canJump=true and
       store landing impact for the landing dip).
  Axis-by-axis naturally yields wall-sliding on diagonal input and prevents corner traps.
- Ramps: solid AABB at full height PLUS a feet-probe each frame; if grounded over a
  ramp's slanted top, snap feet Y to the surface so they don't bounce down slopes.
- Bots use the SAME pipeline with `eye: 0` because bot.pos.y is feet (see P14 + P15).
- Clamp dt to ≤ 0.05 s every frame (see P9).

============================================================
PITFALLS — the failure modes that have actually shipped broken.
Every pitfall lists the wrong forms first, then the correct pattern.
Address each one directly in the code.
============================================================

P1. DISMEMBERMENT-RESPAWN INVISIBILITY (silent killer):
   Wrong: on kill, `scene.attach(part)` to detach the original head/torso/arm/leg meshes
   from the bot group and push them into the debris list. 2 s later debris cleanup
   disposes their material+geometry. On respawn `group.visible = true` — but the parts
   are now backed by disposed material → bot is INVISIBLE while still shooting beams.
   Correct: keep originals under `b.group` forever. On kill, for each part:
       const chunkMat = part.material.clone(); chunkMat.transparent = true;
       const chunk = new THREE.Mesh(part.geometry.clone(), chunkMat);
       chunk.position/quaternion ← part's world transform
       scene.add(chunk); debris.push({ obj: chunk, ... });
   Hide the bot via `group.visible = false`. On respawn, just `group.visible = true`
   and reset part local transforms to their rest poses. Cloning material is required
   because the body material is shared across parts (and bots) — fading opacity on a
   shared material would fade other bots.

P2. SILENT BLUE SCREEN (init crash hidden by clear color):
   The renderer's `clearColor`, the scene background, and the body CSS background ALL
   match the sky color (#1A2230). An init-time JS error that prevents `renderer.render()`
   from ever firing looks IDENTICAL to a working but empty scene. To surface the failure:
     - Add `<div id="err">` (red text on dark, hidden until used).
     - In a plain (non-module) `<script>` at the TOP of `<body>`, BEFORE the importmap
       and module script, install `window.addEventListener('error', ...)` and
       `'unhandledrejection'`. Listening early is what catches module-load failures.
     - On either event, show the panel and append the error stack.

P3. CANVASTEXTURE.CLONE() FRAGILITY:
   `canvasTexture.clone()` reuses the canvas source but the upload-flag handling is
   easy to get wrong — walls end up untextured or pure black on some setups. Instead
   build a fresh `THREE.CanvasTexture` per mesh that needs its own `.repeat`. The
   `makeGridTex(...)` helper should construct its own offscreen canvas every call.

P4. FIRST-FRAME LATENCY / BLANK CANVAS:
   The first `requestAnimationFrame` callback may run several ms after page paint;
   user sees a blank canvas in that window. Call once after building the scene,
   BEFORE the first rAF:
       applyCamera(0);
       renderer.render(scene, camera);

P5. CANVAS SIZING — let CSS own layout:
   Style the canvas with `position: fixed; inset: 0; width: 100vw; height: 100vh;`.
   Call `renderer.setSize(window.innerWidth, window.innerHeight, false)` — the third
   arg `false` keeps three.js from clobbering the CSS size. Re-run setSize + update
   `camera.aspect` + `updateProjectionMatrix()` in a `resize` listener.

P6. RAMP VISUAL — must read at a glance:
   Wrong: solid AABB plus a near-zero-height slab (0.04 u). From the side it looks
   like a paper-thin line on top of a box — invisible from most angles.
   Correct: solid box at full ramp height PLUS a slanted slab (~0.06 u thick) covering
   the top, rotated by `Math.atan2(height, lengthAlongAxis)` and positioned at
   `height/2`. Slab uses a slightly lighter shade than the cover color so the ramp
   pops visually.

P7. POINTER-LOCK PAUSE LOGIC — single source of truth:
   The browser auto-exits pointer lock when the user presses ESC. Handle pause/resume
   PURELY from `pointerlockchange`. Do NOT also bind ESC to a key handler — that
   would double-toggle. Guard `requestPointerLock` against multiple calls in the same
   frame (compare against a frame counter).

P8. RAYCAST USING YOUR OWN AABBs:
   Do not use `THREE.Raycaster` for shooting — it's slower and tangles with the
   non-mesh objects you add. You already maintain `staticAABBs` for collision; reuse
   them with a ray-vs-AABB slab method. Use the same routine for bot bodies. The
   slab method returns the entry axis trivially, which gives you the wall-impact
   normal for the dark quad orientation.

P9. NaN + dt SAFETY (the "where am I?" bug):
   - Clamp `dt` to [0, 0.05] every frame. Without this, a tab-switch sends dt to
     several seconds and tunnels the player out of the world.
   - If any component of `player.vel` or `player.pos` is non-finite, reset that
     value (vel → zero, pos → spawn). Same guard for bots.

P10. AUDIOCONTEXT MUST BE LAZY:
   Browsers block autoplaying AudioContexts. Create the context on the FIRST user
   gesture (the click that grabs pointer lock or fires). Wrap in try/catch; missing
   audio is acceptable, a crashed game is not.

P11. BOT HITBOX = LEGS-ONLY ("I shot him in the chest, nothing happened"):
   bot.pos.y is the FEET position (0 on flat ground). The bot AABB for raycasts
   must span the FULL silhouette:
       min = ( pos.x - BOT_HW_X,  pos.y,            pos.z - BOT_HW_Z )
       max = ( pos.x + BOT_HW_X,  pos.y + BOT_H,    pos.z + BOT_HW_Z )
   Wrong forms that produce the chest-shot bug:
     - `pos.y - h/2 .. pos.y + h/2` (centers on feet → half the box is underground)
     - Using only the torso mesh's bounding box (no head, no legs)
     - Using `BOT_H = 1.0` (only legs)
     - Caching an AABB built at spawn time and never refreshing it
   REBUILD the AABB from the live `bot.pos` on every raycast.

P12. HIT DETECTION ORDER + STALE MATRICES:
   Do NOT call `Raycaster.intersectObjects` on the bot group — its world matrix can
   be stale, and recursion through the part hierarchy is fragile after dismemberment.
   Use `bot.pos` directly and a ray-vs-AABB. Exact loop:
       1) Raycast vs all `staticAABBs` → bestT, bestNormal, hitBot=null.
       2) For each ALIVE bot, ray-vs-AABB its live box.
          If t < bestT: bestT=t; hitBot=b; bestNormal=null.
       3) end = origin + dir * (isFinite(bestT) ? bestT : maxRange).
       4) Single beam from muzzle → end.
       5) if (hitBot) killBot(hitBot, end);
          else if (bestNormal) spawnWallImpact(end, bestNormal);
   Bots and walls must be tested in the SAME pass. Bots take precedence when nearer.

P13. INVERTED OR SLUGGISH WASD:
   Two bugs, one root cause — wrong forward/right math, OR smoothing the input vector.
   CORRECT math (camera at yaw=0 looks down −Z, so "forward" must be −Z):
       forward = ( -sin(yaw), 0, -cos(yaw) )
       right   = (  cos(yaw), 0, -sin(yaw) )
       wish = (W?+forward) + (S?-forward) + (D?+right) + (A?-right)
       if (wish.lengthSq() > 0) wish.normalize()
       desired = wish * (sprint ? SPRINT : WALK)
   Wrong variants that flip the controls:
     - `forward = ( sin(yaw), 0, cos(yaw) )` — W walks BACKWARD
     - `camera.getWorldDirection(...)` — includes pitch; looking up walks you upward
     - Deriving yaw from `camera.quaternion` via atan2 every frame — drift over time
   Apply the clamped-delta acceleration from MOVEMENT. Do NOT lerp `hv` toward
   `desired`, do NOT spring-smooth wish, do NOT lowpass `e.movementX`. GROUND_ACC=90
   IS the responsive feel.

P14. BOTS FLOATING IN THE AIR:
   Bots MUST run gravity AND the collision pipeline every frame. Per alive bot:
       bot.vel.y -= GRAVITY * dt
       resolveBody({ pos: bot.pos, vel: bot.vel,
                     hwX: BOT_HW_X, hwZ: BOT_HW_Z,
                     heightTotal: BOT_H, eye: 0 },     // ← eye=0, bot.pos.y is feet
                   dt, null)
   Failure modes:
     - Spawning bots at `pos.y = 1.0` and never calling resolveBody on them
     - Setting `pos.y = 0` but never integrating gravity
     - Running player collision but skipping bots
     - Using a smaller collision heightTotal than the AABB heightTotal
   Then drive the visual: `bot.group.position.set(bot.pos.x, bot.pos.y, bot.pos.z)`
   — group origin at feet, so part offsets (head ~1.7, torso ~1.05) place correctly.

P15. PLAYER vs BOT COORDINATE ANCHORS (don't confuse):
   - Player `pos.y` = EYE height. ResolveBody body: `eye: EYE_H (1.6)`, `heightTotal: PLAYER_H (1.7)`.
   - Bot    `pos.y` = FEET height. ResolveBody body: `eye: 0`,            `heightTotal: BOT_H (1.9)`.
   The shared `resolveBody` derives an AABB as
       min.y = pos.y - eye         max.y = pos.y - eye + heightTotal
   so passing the right (eye, heightTotal) pair makes it work for both. Mixing them
   (e.g. giving the bot `eye: 1.6`) puts its collision box underground.

P16. MOUSE-LOOK JUMPS / VIEW JOLTS (the "ruckartig" bug):
   Symptom: while looking around, the view occasionally snaps a large angular distance
   in a single frame — feels like a teleport, often after pointer-lock acquisition,
   after ESC→re-lock, after Alt-Tab/refocus, or during fast mouse swipes.
   Five independent root causes — ALL of these must be addressed; fixing one alone
   leaves the bug intermittent:

   a) FIRST-EVENT SPIKE AFTER POINTER-LOCK ACQUIRE:
      The first 1–2 `mousemove` events after `pointerlockchange → locked` can carry
      an outsized `movementX`/`movementY` representing accumulated motion since the
      lock request (Chromium especially). This is the most common cause of "click to
      lock and the view whips around".
      Fix: arm a `dropNextMouseDelta = true` flag inside the `pointerlockchange`
      handler whenever the element becomes the locked target. The next `mousemove`
      with `pointerLockElement === canvas` reads movementX/Y, RESETS the flag, and
      returns WITHOUT updating yaw/pitch. Do this on every (re-)lock, not just the
      first one — ESC + re-lock has the same issue.

   b) OS-LEVEL MOUSE ACCELERATION:
      Windows / macOS apply non-linear pointer acceleration to `movementX`. Fast
      flicks produce wildly larger deltas than slow drags do. Players read this as
      "jumpy".
      Fix: always call `canvas.requestPointerLock({ unadjustedMovement: true })`.
      If the returned promise rejects (older Safari/Firefox), fall back to the
      no-arg form. Do NOT try to "smooth out" acceleration yourself — that adds
      input lag and is the schwabbelig anti-pattern from P13.

   c) PER-EVENT DELTA OUTLIERS:
      Browser bugs, tab-switch return, or graphics-driver stalls can occasionally
      deliver a single `mousemove` with movementX > 500 px. Even with (a) and (b)
      fixed, a stalled tab returning can still emit one nasty event.
      Fix: clamp each delta component to ±MOUSE_MAX_DELTA (use 120) before applying:
          let dx = e.movementX, dy = e.movementY;
          if (dx >  120) dx =  120; else if (dx < -120) dx = -120;
          if (dy >  120) dy =  120; else if (dy < -120) dy = -120;
          yaw -= dx * SENS;  pitch -= dy * SENS;
      120 px at SENS=0.00066 is ~4.5° per event — large but believable. Anything
      bigger is almost certainly an outlier, not a real human flick.

   d) MULTIPLE LISTENERS / WRONG TARGET:
      Two `mousemove` listeners (e.g. one on canvas + one on document) double every
      delta. A listener on `canvas` alone can miss events while the cursor briefly
      sits outside the canvas during lock handoff.
      Fix: exactly ONE `mousemove` listener, on `document`. Search the file before
      shipping to verify there are no others — `addEventListener('mousemove'` must
      appear exactly once.

   e) COALESCED-EVENT DOUBLE-COUNT:
      `event.getCoalescedEvents()` returns sub-events whose `movementX` ALREADY sum
      to the outer event's `movementX`. Iterating sub-events and adding their
      `movementX` on top of the outer one doubles rotation on every flick.
      Fix: use ONLY the outer event's `movementX`/`movementY`. Do NOT call
      `getCoalescedEvents()` for mouse-look.

   Also: NEVER read movementX/Y when `document.pointerLockElement !== canvas` — a
   stray mousemove during the unlocked / paused state would rotate the camera while
   the user isn't looking. The pointer-lock guard is the first line of the handler.

   Anti-fixes that look helpful but make the problem worse:
     - Lerping yaw/pitch toward a target. Adds latency, hides spikes instead of
       suppressing them, makes aim feel mushy. Forbidden by MOUSE LOOK rules.
     - Reading movementX from inside the rAF loop instead of in the event handler.
       Coalesces and lags input; aim drifts behind the cursor.
     - Dividing movementX by `devicePixelRatio`. Pointer-lock deltas are already
       in CSS pixels regardless of DPR; dividing makes high-DPI screens sluggish.

============================================================

CONTROLS:
- WASD move, mouse look, Space jump, Shift sprint
- Left click: railgun shot (also acquires pointer lock if not yet locked); during the
  ELIMINATED overlay, left click forces immediate respawn
- ESC: browser releases pointer lock → game auto-pauses; next click re-locks and resumes
- Simulation is delta-time and frame-rate-independent; rAF natively, no manual fps cap

WEAPON — RAILGUN (only weapon, one-shot-one-kill, hitscan):
- Cooldown RAIL_COOL = 0.4 s. Thin cooldown bar bottom-center.
- Raycast: see P12 for the exact loop. Bot hit = instant kill, always.
- Beam visual: two concentric cylinders muzzle → impact point, AdditiveBlending,
  depthWrite:false, ~250 ms life with opacity fade. Outer ~0.07 radius, accent color,
  opacity ~0.45. Inner ~0.018 radius, white, opacity ~0.9.
- Muzzle flash: brief PointLight (intensity ~8, range ~6, ~100 ms) + small additive
  Sprite at the shot origin
- Wall impact (see HIT FEEDBACK — small version)
- Heavy visual recoil on the camera ONLY: FOV punch (+~6°) + position offset along
  shot direction (~−0.08) + slight up nudge (~+0.025), all springing back over
  ~200 ms. Recoil NEVER alters yaw/pitch.
- WebAudio synth: bandpassed noise burst (~180 ms) + descending sawtooth 880→120 Hz

VISUAL STYLE — ESPORTS-CLEAN, BRIGHT, READABLE (NOT dark, NOT moody):
- "Bright daylight scrim", not "dim corridor". Reads clearly on a laptop in a lit room.
- NO vignette, NO radial darkening, NO color-grade overlays.
- Palette (exact hex):
    Sky / background:    #1A2230
    Floor base:          #4A5360       Floor grid lines:  #6E7986
    Wall base:           #545E6C       Wall grid lines:   #7A8492
    Cover / ramps:       #5E6876
    Accent (beam, HUD, marker):  #B8EEFF
    Bots:                #E86A3C
- Lighting (bright and even):
    HemisphereLight(sky=0xCFE3F5, ground=0x3B424E, intensity=0.85)
    DirectionalLight(0xffffff, 0.95), positioned above-front, soft shadows
      (PCFSoftShadowMap), shadow camera frustum ~±32 in X/Z, near 1, far 80
    AmbientLight(0x404858, 0.25)
- Renderer: ACESFilmicToneMapping, exposure 1.15, outputColorSpace = SRGBColorSpace,
  clearColor = SKY, antialias on, pixelRatio = min(devicePixelRatio, 2)
- Scene fog: `Fog(SKY, 40, 95)` — walls at ~25 u stay crisp; far edges blend out
- Materials matte (MeshStandardMaterial, roughness ~0.85, metalness 0). Only beams,
  bots and HUD accents slightly emissive. Glow faked with emissive + additive sprites
  — no bloom.
- HUD typography: thin system-ui sans-serif, lots of negative space.

MAP (procedural, deterministic):
- Square arena ARENA × ARENA, walls WALL_H tall around the perimeter
- Floor + walls: canvas-generated grid texture (base color + lighter grid lines,
  ~64 px cell on a 512 px canvas, 2 px lines drawn with `+0.5` offset for crisp
  1-px-equivalent strokes). Tile via `texture.repeat`. SRGBColorSpace, RepeatWrapping.
  EVERY textured mesh that needs its own `.repeat` gets its own fresh CanvasTexture
  (see P3).
- Deterministic RNG: mulberry32 seeded with SEED — placement reproducible across runs
- COVER_COUNT cover boxes of varying size scattered, NONE within 4 u of spawn
- RAMPS slanted boxes for height variation (see P6 for the visual)
- All world geometry from `BoxGeometry`, consistent material style

BOTS (3) — TUNED EASY, the player should usually win duels:
- Humanoid silhouette built from box parts grouped under a `THREE.Group`:
    head   (0.36 cube, body material) at y ≈ 1.7
    torso  (0.6 × 0.9 × 0.35, body material) at y ≈ 1.05
    armL/R (0.18 × 0.8 × 0.18, dark material) at y ≈ 1.05, x ±0.45
    legL/R (0.22 × 0.85 × 0.22, dark material) at y ≈ 0.43, x ±0.16
  Body material is orange accent; arm/leg material is dark matte.
- Dimensions for collision and raycast: BOT_HW_X, BOT_HW_Z, BOT_H (see Constants).
- AI loop: wander toward random points at BOT_SPEED; raycast LOS to player; if LOS
  held for BOT_LOS_TIME continuously AND within BOT_RANGE, fire on next ready tick.
- Fire cooldown: BOT_FIRE_COOL + rand() * BOT_FIRE_JIT.
- Aim spread: perturb the aim direction by random amounts in the screen-space basis
  around the aim direction (right, localUp), each scaled by BOT_SPREAD radians.
- Visible orange beam, same primitives as the player beam.
- One-shot-one-kill both ways.
- Respawn after BOT_RESPAWN at a point ≥BOT_RESPAWN_MIN_DIST from the player and not
  inside any static AABB.
- See P1 — respawn restores the ORIGINAL part meshes; debris is made from CLONES.
- See P14 — bots run gravity + resolveBody every frame.

HIT FEEDBACK — must feel juicy (this is the core of the game):
Every player-fired hit on a bot triggers ALL of the following, layered:

  Visual on the target:
  - Bot group hides; CLONED chunks of head/torso/arms/legs spawn as debris (see P1)
  - Each chunk: radial initial velocity from impact + slight upward bias, random
    angular velocity, gravity ~16 u/s², ground bounce at y ≈ 0.08 with damping
    (~0.32 restitution, ~0.6 horizontal friction)
  - ~40 small additive shard particles (tiny boxes, bot accent + white) fly out
    radially and fade
  - Bright white additive Sprite flash at impact (~80 ms)
  - Expanding ground shockwave ring (RingGeometry, accent color, additive-ish
    opacity, grows ~5× over 400 ms)
  - Brief upward dust puff (soft sprites, fades over ~600 ms)
  - Debris lifetime ~2 s, then fade + dispose CLONES' material/geometry. NEVER
    dispose the originals.
  - NO blood, NO gore — clean geometric fragmentation only

  Camera/screen feedback:
  - Tiny camera punch toward the target (transient position offset only — NEVER
    yaw/pitch)
  - Crosshair hit-marker: 4 short cyan strokes flash outward from center for ~120 ms

  Audio (all WebAudio synth, see P10 for lazy init):
  - Sharp high-frequency "tink" (triangle 2.4k → 0.9k, ~120 ms)
  - Layered low sub-thump (sine 120 → 45, ~200 ms)
  - High-passed noise sweep for the shatter
  - ±10% random pitch variation per hit

  Game-state:
  - Score counter scales up briefly when incrementing (CSS transform pop)

Wall hits get a smaller version: ~14 spark particles, small dark quad oriented to
the hit-surface normal and offset 0.01 along the normal to avoid z-fighting (no
DecalGeometry), soft impact sound. No camera punch, no hit-marker.

PLAYER DEATH & RESPAWN:
- Camera position jolt (not rotation), screen fades to dark slate with thin cyan
  "ELIMINATED" text, 1.5 s auto-respawn
- LEFT CLICK during the ELIMINATED screen forces immediate respawn
- Respawn at the spawn point, velocity zeroed, HP=100; yaw/pitch preserved; bots not reset

HUD (DOM overlay over the canvas, `pointer-events: none`):
- Crosshair: fine 1 px + in accent cyan with thin dark outline
  (`box-shadow: 0 0 0 1px rgba(0,0,0,0.55)`) for readability on any background
- HP bottom-left (number + thin bar)
- Score bottom-right (kills) with the scale-pop on increment
- Cooldown bar bottom-center, thin, accent-colored, fills as the railgun recharges
- Pause overlay: semi-transparent dark layer + centered thin "PAUSED"
- Eliminated overlay: darker layer + centered thin "ELIMINATED"
- Error overlay (#err): hidden until P2 catches something

============================================================
CODE STRUCTURE:
============================================================
- One `<script type="module">` block for the game.
  Plus a tiny non-module `<script>` at the very top of `<body>` that wires the
  error panel (see P2) — must come BEFORE the importmap so it catches
  module-load failures.
- Clear sections in this order:
    constants (the Constants block above, literally)
    RNG (mulberry32)
    renderer / scene / camera, resize listener
    lights
    grid textures (makeGridTex)
    world (addBox, addRamp, floor, walls, cover, ramps)
    player state
    input (keyboard, mouse, pointer-lock; AudioContext lazy-create)
    collision (rayAABB, rayAABBNormal, resolveBody)
    updatePlayer, applyCamera
    audio (railSound, hitSound, wallSound)
    bots (buildBot, spawnBots, respawnBot, reassembleBot, updateBots, botFire,
          hasLOS, pickPointAwayFrom)
    railgun (fireRailgun)
    transient FX (spawnBeam, spawnMuzzleFlash, spawnWallImpact)
    killBot (clone chunks → debris; see P1)
    debris + transient + ring update
    HUD glue, score/HP/death/respawn
    initial render (see P4)
    gameLoop + first rAF
- Frame-rate-independent (delta time everywhere)
- Comments only where genuinely needed. Especially: mark the dismemberment-clone
  pattern (P1) so a future maintainer doesn't "simplify" it back into the broken form.

DELIVERABLE: only the full content of `game.html`, nothing else.
Ignore all existing files and structures, only focus on generating `game.html`, nothing else.
