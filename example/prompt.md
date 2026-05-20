# Procedural 3D Galaxy — Single-File Browser Experience

Build me one self-contained HTML file called `galaxy.html` that opens directly in a modern desktop browser and runs a procedurally generated 3D space simulator the user can fly through. Treat this as a finished piece of work, not a tech demo — the first frame must feel cinematic, the controls must feel good, the audio must feel alive. Beauty and restraint matter more than feature count: if at any point you can't decide between adding more or polishing what's there, polish.

You are acting as a senior creative-coding engineer with strong WebGL, Three.js, and procedural-content experience. Write the file as something you would be proud to ship.

---

## Tech constraints

- Exactly one file. No build step, no bundler, no companion files, no external textures, no model files, no audio files.
- HTML with `<style>` inline and a single `<script type="module">` block containing all logic.
- Use Three.js loaded from a CDN through an import map, together with the matching post-processing add-ons: an effect composer, a render pass, an unreal-bloom pass, and a custom shader pass for the final grade. Do not pull in any other library.
- The page must run in current Chrome, Firefox, and Safari without warnings or runtime errors. No console errors on load, no console errors during normal play.
- Target around 60 fps at 1080p on a mid-range integrated GPU.

---

## Opening moment

When the page loads the camera is already inside the universe, slowly orbiting a warm yellow home star. A large title overlay reads `GALAXY`, with a small wide-letterspaced subtitle below it along the lines of `A PROCEDURAL UNIVERSE`. A discreet button or label in the lower half says something like `► CLICK TO START`. Both overlays sit in front of the 3D scene with soft green-neon styling and a faint glow shadow.

Any click anywhere on the page must:

1. Initialize the Web Audio context (browsers refuse to start audio before a user gesture).
2. Request pointer lock on the document body.
3. Hide the title and the start label.
4. Stop the cinematic intro orbit and hand the player full control.

When the player presses `Escape` and pointer lock is released, show the start label again as `► CLICK TO RESUME`. Another click re-acquires pointer lock and resumes play. Don't show the giant title a second time — just the resume label.

The intro orbit itself should be a slow drifting arc around the home star with a tiny vertical bob, framed so that one specific hero body (described below) is already on screen. The framing has to read as a polished establishing shot — symmetrical-feeling, calm, no jitter.

The starting camera position is offset from the home star far enough that the star reads as a sun, not a wall of light. Roughly one to two star radii out and a small distance above the orbital plane is a good ballpark. The camera should look slightly past the star rather than directly at it, so the hero planet has center-frame real estate. Movement during the intro is gentle — angular speed around the star is slow enough that nothing whips by; the vertical bob is small enough that it feels like camera breathing, not motion sickness.

---

## Ship controls and feel

Camera orientation is a quaternion you accumulate yourself — never lean on Euler angles or `lookAt` for control. The player must be able to roll fully upside down at any pitch with no gimbal pop and no horizon snap.

Bindings:

- `W` and `S` — forward and reverse thrust.
- `A` and `D` — lateral strafe.
- `Q` and `E` — roll left and right.
- Mouse — yaw and pitch (pointer-lock based, relative movement).
- `Space` — hard brake.
- Left mouse button — fire forward laser.
- `H` — hide and reshow the HUD and crosshair.

Mouse look should feel snappy but not raw. Smooth the raw mouse deltas with a short exponential time constant so micro-jitter from a high-DPI sensor is absorbed without adding noticeable lag. Cap the per-event delta defensively in case the OS delivers a huge jump after window focus changes. Right after pointer lock re-acquires, drop the first few mouse events — browsers sometimes feed a stale large delta on lock change.

Throttle behavior is the central feel knob. Get this right:

- Holding `W` or `S` follows a non-linear acceleration curve. The first short window is an eased ramp-in that reaches roughly a quarter of top speed, then a longer ramp lifts that toward full top speed. The intent: a quick press taps you a small amount forward, but holding the key longer commits to faster and faster travel. Top speed is around thirty times the in-scene "speed of light" unit (a small named constant, used purely as a HUD readout unit — there is no actual physics here).
- Releasing `W` or `S` preserves the current speed. Space has no friction; the ship coasts indefinitely.
- Re-pressing thrust mid-coast must continue from the current speed fraction along the same curve, not snap back to zero. The cleanest way is to invert the curve from the current speed and continue the hold timer from there.
- `Space` is a hard exponential brake that brings the ship to a full stop in about a fifth of a second. It also plays a distinct brake sound. After the stop, holding `Space` does nothing further.
- Strafing scales mildly with current forward speed — at rest you strafe slowly, at high speed you slip sideways faster.
- When the camera gets close to a planet or moon surface, automatically and quietly bleed off speed so the player physically cannot pile-drive through it. Below a small clearance, decay aggressively; just above that, decay gently. This is a comfort feature, not a collision system.

Frame deltas must be clamped (somewhere around 50 ms max) so a tab-switch or hitched frame doesn't teleport the ship into the next galaxy.

A note on scale and feel: this universe is huge, so the player should feel speed even when relatively far from anything. The non-linear throttle curve plus the in-scene speed-of-light readout exists for exactly this reason — when the HUD shows several times the speed of light, the player should believe it. Pick scene-unit sizes accordingly. The home star is on the order of a few hundred units in radius, planets a few tens of units, the loaded chunk volume is hundreds of thousands of units across. The back skydome sits well past the loaded chunk volume so it cannot be approached.

---

## HUD

The HUD lives in the top-left corner. Use a small monospace font, terminal-green tint, a soft text-shadow glow, and slightly wide letter-spacing so it reads as a holographic overlay. The HUD has exactly two lines:

1. A subdued reminder of the bindings — something like: `W/S thrust · A/D strafe · Q/E roll · space brake · LMB fire · H hide`.
2. A live status line that shows:
   - Current speed as a fraction of the in-scene speed-of-light unit, with a direction marker (`▶` forward, `◀` reverse, a dot at rest), and a small brake indicator when the brake is active.
   - The nearest body — its class (stellar class letter for a star, or `gas` / `ice` / `rocky` for a planet, etc.) and its distance from the camera in scene units, accounting for the body's radius so the number is "gap to surface" not "gap to center".

A fixed `+` crosshair sits at the screen center in the same neon green with a soft glow. Both the HUD and the crosshair toggle together via `H`.

Style note: this is a heads-up display, not a UI panel. Keep it sparse, single-color, glowing. Resist the urge to add bars, icons, or boxes.

---

## Universe — infinite, chunk-streamed, deterministic

Space is unbounded. Subdivide it into a 3D grid of cubic chunks identified by integer coordinates. A chunk's contents are determined entirely by its integer coordinates through a stable seeded pseudo-random generator — visiting the same coordinate twice always produces exactly the same stars. Use a small fast 32-bit hash-style PRNG (mulberry32 or similar) seeded by mixing the chunk coordinates with prime multipliers, so the seed feels truly mixed.

Keep a generously sized cube of chunks live around the camera at all times — large enough that thousands of stars are simultaneously active and the player never sees a hard pop on the horizon. Rebuild the active set only when the camera crosses a chunk boundary, not every frame. When a star scrolls out, drop it from the active list; cache the chunk's contents so re-entering it later is instant and identical.

Above the chunk grid, layer a coarser "region" grid (each region a small cube of chunks). Each region picks a theme from a coarser hash. Roughly four themes, with hand-tuned probabilities:

- **void** — rare or zero stars, but the ones that do exist are very large.
- **cluster** — many small stars tightly bunched around a single anchor inside the chunk, suggesting an open cluster.
- **normal** — the default. One or two stars at random offsets.
- **giants** — uncommon, one or two very bright hot stars biased toward the top of the size range.

The chunk at the universe origin always contains one star — the home star — and that star is always a yellow G-class so the opening frame is reliably warm. The home star's identity, color, and radius must be deterministic so the rest of the file can reference it directly during scene setup.

Track the nearest non-dead star to the camera every few frames. When the nearest star changes, lazily build the new star's planetary system and tear the previous one down cleanly so memory does not grow over time.

---

## Star rendering

A star is rendered in two cooperating layers:

- **Far layer — point sprites.** Every star in the loaded volume is drawn as a single GL point with a custom shader. Each point has a sharp Gaussian core in the star's stellar-class color. The projected pixel size depends on physical radius divided by distance; clamp it so a very close star can't balloon. Brightness is multiplied by a slow per-star twinkle (a single low-amplitude sine using a stable per-position seed — a few percent of amplitude, not a flicker). For the rare close large points, fade in a faint cross-shaped diffraction-spike pattern; do not paint spikes on the small far ones. Stars must read as clean small discs, never as glowing pillows.
- **Near layer — additive glow billboards.** Every star also exists as an instanced camera-facing billboard plane with a richer shader: a tight bright core, a softer mid glow, a wide faint halo, vertical and horizontal diffraction spikes, and a fainter forty-five-degree spike pair. The billboard's world-space size scales with the star's physical radius and includes a subtle per-star pulse so different stars breathe at different rates.

The two layers should feel like one star. The point-sprite layer is the dominant look at distance; the glow billboard takes over for stars that are physically nearby. Twinkle on the points and pulse on the billboards should not synchronize, so the sky never feels like a coordinated blink.

Each star also has a corresponding invisible 3D sphere mesh, sized larger than the visible glow, registered for raycasting only — the laser needs a real hit target per star so it can identify which one was shot. Use one instanced mesh for all stars; carry an index-to-star table so a hit instance ID can be resolved back to a star object.

Stellar classes drive color:

- M (cool red) — most common.
- K (orange) — common.
- G (yellow-white) — average.
- F (white) — rarer.
- A (blue-white) — rare and only at large radii in giant regions.

Brightness discipline matters. Caps that have worked well in practice: hard upper bound on apparent point size, per-point output multiplier below one, modest tone-mapping exposure, and bloom that bites only the brightest pixels. Bloom plus additive plus uncapped sprite size will turn the entire field into white mush — don't.

A bit more about how each stellar class should read on screen:

- M (red dwarf) — small, dim, warm orange-red. The most common star in the universe and should be visually unremarkable from a distance. Up close still small and warm.
- K (orange) — medium-small, warmer than M, more luminous. The everyday star.
- G (yellow) — neutral warm white-yellow. The home star is one of these. Calm, balanced, the visual anchor of the opening shot.
- F (white) — bright pure white. Less common, more striking.
- A (blue-white) — large, bright, cool blue-white. Rare, used as a visual punctuation mark in the universe. When the player flies past one it should feel different.

Pick the per-class color tints carefully and stick to them. Resist the temptation to over-saturate — astronomical-photo-like restraint reads more believable than candy-colored stars. Then per-star, multiply that base tint by a small random factor so two M stars next to each other aren't identical.

A single real-time directional light tracks whichever star is currently nearest. Its color matches that star's tint, its position is the star's position, and its target follows the camera so planets always have a correct sun-side terminator.

When the player is flying through interstellar space between systems, the directional light should keep referencing the closest star — and only the closest. Multiple suns lighting one scene at once look wrong here; this is intentionally a one-sun-at-a-time piece. The transition when the nearest star changes does not need to be smooth — by then the previous system is being torn down anyway, and the visual swap reads as part of arriving at a new place.

---

## Background sky

Two layers, both parented to a group whose translation is locked to the camera every frame. The group rotates as the world rotates, but it does not translate as the player flies, so the sky reads as infinite — only the player's rotation moves it past, never their position.

- **Procedural deep-space skydome.** An inverted sphere painted by a fragment shader. The base is deep navy that brightens very slightly toward "up" so the void never reads as flat black. Over that, layer a procedural Milky Way galactic band cutting diagonally across the sky: broad soft FBM-noise cloud puffs along the band, a few darker dust silhouettes carved into them, and a warmer brighter galactic-core bulge concentrated in one direction. The band is anchored in world space; flying around must not make it shift relative to itself. Use a fixed galactic-pole direction defined as a constant in the shader so the band's tilt is stable across runs. The bulge sits along its own constant direction so the player can always orient themselves by "the warm side of the band".
- **Background pinprick stars.** Around seven thousand small dots distributed uniformly on a sphere out at the back distance. Use the same stellar-class palette as the foreground but with a duller multiplier so they read as faint and small. Plus a sparse second layer of a few dozen brighter "anchor" stars biased toward the galactic plane — same palette, slightly larger, with a very subtle twinkle and diffraction spikes only on the rare largest ones.

These layers together fill every direction so the sky never feels empty regardless of where the streamed near stars happen to be.

---

## Planetary systems

When the player approaches a star, build its system: four to eight planets at deterministically seeded orbital radii, eccentricities, and inclinations. Inclinations must genuinely vary in 3D — orbits should not all share one plane. Build each system lazily on approach, dispose it cleanly on departure. Disposing means removing meshes, releasing geometries, releasing materials, releasing instanced fields, and unregistering anything from the destructible list.

Three planet types, picked with a realistic bias toward rocky and gas:

- **Rocky** — small to medium. Procedurally paint the surface into per-vertex colors using deterministic position-based noise. Mix continents and oceans and deserts with polar ice caps in the highest latitudes. About a third of rocky planets are ocean-less arid worlds. Some carry a thin atmosphere shell, others don't.
- **Ice** — cool pale blues and whites with frosty subtle latitudinal banding. Most carry a thin bluish atmosphere.
- **Gas giant** — large, with latitudinal bands painted into vertex colors from one of several Jovian-style palettes (warm ochre, cold blue-white, deep indigo, salmon and rust, pale lavender). Add storm-like darker patches near the equator. Every gas giant gets a thicker atmosphere shell and most carry an animated cloud layer painted onto a slightly larger translucent sphere via FBM noise that drifts over time.

Every planet's atmosphere shell renders as a colored back-side rim glow with a sun-facing brightness bias and a low-amplitude shimmer animated over time so the atmosphere quietly breathes. Some planets get a subtle eclipse-side halo enhancement so the silhouette catches the eye even at low light.

Many planets carry moons of varied flavors — grey rocky, icy white, volcanic ochre with a faint emissive glow, rust-red, bluish ocean. Each moon orbits on its own randomly tilted plane around its parent, with its own phase and angular speed.

Ringed planets get two cooperating systems:

- A textured banded ring disc with several gaps drawn into the texture (canvas-painted with multi-frequency sinusoidal bands, plus a few sharp gap notches). The ring shader respects the planet's shadow — the ring darkens where the planet would block the sun.
- A real instanced field of small 3D asteroid debris embedded inside the ring's volume, orbiting around the planet's spin axis. Each debris instance also has its own slow individual tumbling rotation around a random axis, implemented in the vertex shader via a per-instance axis+angular-speed attribute so the whole field is still one draw call.

Different planets carry differently tilted rings — never all aligned to the orbital plane.

Orbits, axial spin, ring rotation, atmosphere shimmer, and cloud animation all advance with time.

A planet's vertex-painted surface is what carries the close-up. Spend the noise budget on it. For a rocky planet that means at least two layers of position-based noise mixed at different scales — large continents from low-frequency noise, ridges and basins from higher-frequency noise — together with latitude-dependent ice caps and biome-dependent palettes. For a gas giant that means clearly visible latitudinal bands from a primary cosine wave biased by latitude, plus a sharper second band layer for finer detail, plus the storm-like darker patches near the equator picked by yet another noise lookup. For an ice planet a quieter pair of noise layers producing soft mottled white and pale blue with darker veins is enough.

The atmosphere shell is rendered as a slightly larger sphere with back-side faces only, additive blending, no depth write. The shader does a Fresnel rim using the angle between the surface normal and the view direction, brightened on the side facing the active star and dimmed on the night side. A subtle low-amplitude shimmer driven by a stable per-planet seed and the running time gives the atmosphere life. Avoid making the rim too thick — a thin glow on the silhouette reads as atmosphere; a thick glow reads as a halo.

The cloud shell on gas giants is rendered just slightly larger than the planet sphere, with normal blending, no depth write. The shader paints animated FBM noise into a banded latitudinal pattern. Drift its phase slowly with time so the bands clearly move as the player watches.

---

## Showcase planet

Inside the home system at the universe origin, force one planet to be a clearly showcase-quality body. A large ringed gas giant with:

- Detailed banded rings that fill most of its frame, with visible gaps, a warm sandy tint, and the planet's real shadow falling across them.
- A dense visible field of instanced asteroid debris inside those rings, slowly rotating around the planet's axis.
- Three contrasting moons in clearly differentiated orbits — for example an ochre volcanic one with a faint emissive glow, an icy white one, and a smaller blue ocean one.
- A vibrant warm atmosphere with a slight eclipse-side halo.
- A cloud band layer.

Position this planet's orbital phase and orbital plane so it is already on-screen during the intro orbit. The very first frame the player ever sees should make them stop and look. This is the calling card for the entire piece.

Don't try to make every planet in every system this polished — the other planets are procedural variety, and the contrast between them and the hero matters. The showcase exists so that the opening moment lands; the rest of the galaxy can be more rough-edged and still feel right.

---

## Asteroids

Pre-bake a small set of irregular asteroid geometries — around ten variants. Start from an icosahedron with one or two subdivisions and displace each vertex outward by a deterministic position-based hash noise so neighboring vertices at the same world position move identically and the mesh stays watertight. Share these geometries across the entire world via instanced meshes — never one mesh per rock. Provide several distinct rock materials in different earthy tones, all rough, flat-shaded, non-metallic.

Use these for two purposes:

- **Standalone asteroid belts** inside planetary systems. Each system gets two to five belts. Each belt picks its own random plane normal so different belts tilt differently. Belt inner radius, outer radius, thickness, density, and rock material all vary per belt. Sprinkle a wide distribution of sizes — mostly small pebbles, some medium chunks, a few rare large boulders that read as recognizable landmarks when flown past.
- **Ring debris** inside planetary rings. Same geometry pool, smaller scale distribution, with the per-instance tumbling rotation described above.

---

## Nebulae

Scatter a small handful of distant nebulae across the local region — five is a good count, distributed reasonably uniformly on a sphere around the home system. Each nebula is a soft semi-transparent stellar-nursery cloud, built from a procedurally painted canvas texture used as additive sprite billboards.

The texture should mix:

- A wide soft outer halo in a steel-blue or violet "shell" color.
- Several brighter core blobs in a warm emission color (pink-red, orange, magenta, pale yellow, teal, depending on the scheme).
- A scattering of small bright pinprick highlights suggesting newborn stars embedded in the cloud.
- A few darker punched-out blobs to suggest dust silhouettes.
- A radial mask so the edges fade out softly to fully transparent.

Use a small library of evocative color schemes — pink core with steel-blue shell, ember orange with deep blue shell, salmon with violet, magenta with midnight, emerald with royal blue. Render each nebula as two stacked sprite layers at slightly different scales for a sense of depth, plus a few additional small bright sprites for the newborn stars clustered near the core.

Nebulae are visual landmarks only — no collision, no interactivity, no audio. Their job is to give the player something to fly toward across an otherwise empty sky.

Place them at a distance comparable to the back-skydome scale, not at chunk-streamed distance. That way they read as far-away landmarks rather than nearby objects the player can reach inside a few seconds of thrust. The brightness should be low enough that a nebula does not dominate any frame — it's a horizon element, not a hero.

---

## Black hole landmark

Place one designated black hole near the home star — not at the home position, but reachable from it within a short flight. Build it from four parts:

- A tilted accretion-disc plane mesh, textured with a procedurally painted swirling disc canvas — hot white near the inner edge, transitioning through gold and orange to deep magenta at the rim, with a punched-out central hole and a softly faded outer ring.
- A second disc plane oriented roughly perpendicular to the first at half opacity, suggesting the gravitationally lensed back of the disc looping over the top.
- A camera-facing sprite glow ring tinted warm gold for the photon sphere.
- A pure black camera-facing sprite on top of the lot with elevated render order so it cleanly punches the event horizon over everything behind it.

The bloom pass should make the disc visible from a great distance.

The black hole has no system, no planets, no audio. Its job is to be a landmark that catches the player's eye in flight and pulls them off course. Resist the urge to add gravitational lensing or particle infall — that's a different scope of effect, and a clean static composite reads better here than a half-implemented physics gimmick.

---

## Combat — laser and destruction

Left mouse button fires a thin additively blended green laser bolt that travels forward from the camera and hits the first valid target along its ray. Use a single reusable Raycaster — do not allocate per shot. The laser geometry itself is a thin elongated box parented to the camera, paired with a small soft green glow sprite at the muzzle. Both fade out over a short visible window after each shot.

Targets are stars (via the invisible instanced sphere mesh), planets, moons, and asteroids (via a destructibles list maintained as bodies enter and leave active systems).

Add a small aim-assist cone for stars specifically. If a precise raycast misses every target, but a star happens to be inside a narrow cone in front of the camera and within laser range, snap the shot to that star. This is important: at high speeds the player can't reliably hit a distant star with a perfect ray, and "every star feels unhittable" kills the gameplay loop. Keep the cone narrow enough that the assist is invisible at low speeds.

Each shot triggers a brief laser sound (see audio).

Hits on planets, moons, and asteroids spawn a multi-phase explosion built from cleanly separable parts whose sizes are clamped to sensible upper bounds — a destroyed gas giant must not blast a city-sized white sphere across the camera. Phases:

- A short bright additive flash sphere expanding quickly and fading.
- An expanding camera-facing additive ring shockwave.
- A soft expanding spherical shock bubble. Render its sphere as front-side only so the player inside the bubble is not blinded.
- A burst of large tumbling fragments using the pre-baked asteroid meshes with an emissive material, scattering outward with random axes of spin and individual velocities.
- A denser cloud of smaller faster fragments behind them.
- A radial spark burst as a `THREE.Points` cloud with per-vertex velocities, biased into a few spike directions for a "frozen explosion" silhouette rather than a uniform ball.
- A handful of expanding line-segment ray streaks where the head shoots outward fast and the tail follows with a small delay, giving real motion-blur–like streaks.

After the burst, the target disposes cleanly. Destroying a planet cascades the cleanup to its atmosphere shell, cloud shell, ring disc, ring debris field, and all of its moons. Destroying a moon removes only that moon. Cap the number of live explosions; when a new one would push past the cap, retire the oldest cleanly.

Hits on stars are dramatic — trigger a **supernova**. Same building blocks but larger, longer-lived, and multi-colored:

- Three stacked bright flashes in white, warm gold, and cool blue at different sizes and durations.
- Three stacked shock bubbles in cool blue, warm orange, and pale gold.
- Five concentric shock rings in different stellar colors expanding at different rates.
- A heavier cloud of large tumbling fragments and an even denser cloud of medium fragments, all moving fast.
- A bright primary spark burst.
- Two ray streak bursts at different scales — a big bright warm one and a smaller violet one.
- A long-lived dim violet remnant scattered out as slow-drifting sparks (use a sparks cloud rather than a sphere so the player flying through it later is not blinded).

After the supernova, the star is dead. Mark it dead, hide its glow billboard and its invisible hit sphere by zeroing their instance matrices, mark its slot in the index-to-star table empty, force a rebuild of the chunk-streamed star point geometry so the dead star is also gone from the far pixel layer, and tear down its planetary system if it was the active one.

Play a deep, long, distorted supernova sound (see audio).

A supernova is the dramatic centerpiece of the combat loop, so do not under-cook it. The total event including the lingering remnant should last several seconds. The first second is a violent triple flash and the leading shock bubble. The next few seconds are the rings and debris spreading outward. The tail is the slow violet remnant drifting. The combination of audio depth, multi-color rings, ray streaks, and remnant sparks is what sells it — leave out any one of those and the moment falls flat.

---

## Audio — procedural, inline

All sound is synthesized with the Web Audio API at runtime. There are no audio files anywhere.

Build a small set of independent voices, all routed through a master gain at a modest level:

- **Pad drone.** A continuous low-frequency atmospheric bed — a few detuned sawtooth oscillators at low notes, lowpass-filtered hard so only the warm body comes through. Plays from the moment audio initializes and never stops. Keep its level low so it sits underneath everything.
- **Engine.** A continuous voice that responds to current throttle. Several layers feeding one gain stage: two detuned sawtooth oscillators at the fundamental for a "fat" body, an octave-up square oscillator for bite, a lowpass filter whose cutoff opens with throttle so the voice gets brighter as the player pushes it, a waveshaper for warm clipping, a high triangle "whine" that only fades in past about thirty percent throttle, a band-passed looping noise voice for airflow turbulence, and a slow sine LFO summed onto the master engine gain for a tremolo wobble. The fundamental pitch should shift with throttle so the engine clearly rises and falls as the ship accelerates and decelerates.
- **Laser shot.** A short crackly filtered noise burst plus a fast descending sawtooth sweep. Tens of milliseconds long, bright, distinct, never overlapping with itself in an unpleasant way at rapid fire.
- **Regular explosion.** A muffled punchy thump — a noise burst with a lowpass sweep down from bright to deep, a sub-sine pitch drop layered underneath, and a handful of small random crackles scheduled across the next half-second.
- **Supernova.** A long, deep, distorted wash with a heavy sub-bass whomp at the start and a longer rumble tail. A pair of overlapping low sine pitch drops at different envelopes for body. Crackle field scattered over a couple of seconds. Heavier waveshaper distortion than the regular explosion.
- **Brake.** A short downward filtered noise sweep plus a low sine pitch drop and a small high sawtooth chirp for "rasp", triggered once when the player presses `Space`.
- **Thrust-on.** A short upward filtered noise sweep plus a quick low sine sweep, triggered once when the player presses `W` or `S` from near-rest. Suppress it if speed is already high so the player doesn't get a chirp on every key tap mid-flight.

The audio context must be created and resumed only on the first user gesture — pointer-lock click. Browsers will refuse to start it earlier.

Voicing notes: the engine voice is the one the player will hear constantly, so spend time on its character. It should feel like a turbine spinning up, not a synth pad — that means real harmonics from the square octave layer, real grit from the waveshaper, and real motion from the LFO tremolo. The pad drone should be felt more than heard; if a listener notices the pad as a discrete musical element, it is too loud. The supernova should feel physical — the sub-bass whomp should rattle the desk, not just play in the ears.

---

## Post-processing and color grade

Render through an effect composer with three passes:

- The render pass.
- A tightly tuned bloom pass: a high threshold so only real light sources bloom, a moderate radius, a controlled strength. Bloom should feel disciplined, not smeared.
- A final custom shader pass that does a mild contrast boost, a split-tone (slightly warm in highlights, slightly cool in shadows), and a soft circular vignette. The image should read as a film frame, not a sterile WebGL render.

Configure the renderer for ACES filmic tone mapping at a modest exposure (deliberately low — bright sources should bloom, not white-clip), sRGB output, and a logarithmic depth buffer. The scene spans many orders of magnitude in scale and a linear depth buffer will z-fight badly. Any custom shaders you write that participate in the world depth must include the matching log-depth shader-chunk includes so transparent surfaces compose correctly with everything else.

Lighting is intentionally high-contrast deep-space: a very small ambient light just enough to keep shadowed surfaces from going pure black, plus a single bright directional light tracking the active star. Do not add fill lights — the hard one-directional shadowing is the look.

---

## Engineering quality bar

- **No allocations in the hot loop.** The per-frame animate function and anything it calls must not allocate new Three.js objects or arrays. Declare a small pool of reusable scratch vectors, quaternions, matrices, and colors at module scope and reuse them.
- **Frustum culling.** Disable frustum culling on anything where the bounding sphere is unreliable or the geometry is gigantic — point clouds, sky dome, instanced star meshes, background star groups. Otherwise leave defaults in place.
- **Determinism.** All procedural content is seeded by integer coordinates through the stable PRNG. Two visits to the same place produce the same result, every session.
- **Resize.** Handle window resize for the renderer, the composer, the bloom pass, and the camera aspect ratio.
- **Cleanup.** When a planetary system or destructible is torn down, dispose its geometries and materials and remove it from every list that referenced it. No leaks across long play sessions.
- **User-gesture safety.** Pointer lock and audio context only on a real click, never on load.
- **Sectioning.** Use clear section banners as comments inside the script (renderer setup, star streaming, far star points, background sky, asteroid geometry, planet textures, planets, systems, nebulae, black holes, laser, explosions, audio, controls, HUD, post-processing, render loop). The file is long; a reader should be able to scroll to the section they need.
- **Comments explain why, not what.** Reserve comments for the non-obvious tuning choices — why a clamp value is what it is, why a layer is rendered front-side-only, why a specific seed was picked for the home star. Do not narrate every line.
- **No undefined identifiers.** Anything you call must be defined in the file. Define a small set of math helpers (lerp, clamp, smoothstep, deterministic noise, basis from a random direction) once at the top and reuse them everywhere.
- **Test hooks (optional).** Expose a small object on `window` that lets a test driver query the home star, fire the laser, trigger a supernova at a given point, or stop the intro orbit. Keep it minimal and clearly scoped.

---

## Visual quality bar

The frame the player sees right after pressing start must look like a polished cinematic shot:

- The home star reads as a real glowing body — clean, calm, with a soft spike pattern and no faceted edges.
- The showcase planet drifts in frame with clearly visible banded rings, real 3D debris embedded in those rings, and at least one contrasting moon visible nearby.
- The Milky Way arcs diagonally across the sky with the warm galactic-core bulge off in one direction and dust silhouettes carving through the band.
- A nebula or two is visible somewhere on the horizon, providing a colored landmark to fly toward.
- The black hole is far enough away to be a small but unmistakable curiosity.
- The background star field is dense but disciplined — small, clean, calm, with a handful of slightly brighter anchor stars.
- Bloom carries the highlights, vignette darkens the corners just enough, the whole frame is tonally cohesive.

Common ways this look goes wrong, and what to do instead:

- Stars rendered as huge fluffy glows with no defined core — instead, sharp small core with controlled outer halo.
- Bloom turning the field into white mush — instead, high bloom threshold and modest tone-map exposure.
- Atmosphere shells leaking color across the whole sphere — instead, true Fresnel rim with day-side bias.
- Rings as a flat untextured strip — instead, banded canvas texture with visible gaps, plus real 3D debris on top.
- All orbits in one plane — instead, randomized orbital plane normals per planet and per moon.
- Explosions clipping the camera with a giant white sphere — instead, front-side-only shock bubble and clamped sizes.
- Audio that's either silent or harsh — instead, a quiet pad bed and engine voice always playing, with brief louder events on top.
- All planets in one orbital plane like a flat solar-system poster — instead, randomized inclinations so the system reads as three-dimensional from any approach.
- A black hole built as a torus or a sphere — instead, the layered flat-plane disc + perpendicular second disc + glow ring + black-sprite punch described above. Torus geometry will not look right under bloom and additive blending.
- Stars that all twinkle in sync because the seed source is global time — instead, every star's twinkle phase is driven by a stable per-position hash so each one has its own rhythm.
- An "infinite universe" that is actually one ring of stars looped around the camera — instead, real chunk-coordinate-keyed generation so flying in a straight line for minutes keeps yielding genuinely new chunks.

---

## The opening thirty seconds — a guided walkthrough

This is the experience that has to land. Treat the rest of the prompt as the specification and this section as the acceptance test you write the spec against.

Second zero. The page finishes loading. Title sits center-screen reading `GALAXY` in big spaced-out neon green letters with a soft glow. Subtitle beneath. The camera is already deep in space, slowly arcing around the warm yellow home star. The hero ringed gas giant is in the right half of the frame, its banded rings clearly visible, one moon catching light beside it. A nebula glows faintly in the distance behind. The Milky Way arcs diagonally across the sky. A handful of brighter background stars punctuate the dark.

Second three. The player notices the `► CLICK TO START` label at the bottom and clicks. Pointer lock engages. The title fades. The intro orbit stops. The HUD appears in the top-left in matching neon green, showing the bindings and the speed status as zero. The audio context starts; the pad drone fades in low and warm, the engine voice settles at idle.

Second five. The player moves the mouse. The camera turns smoothly without acceleration weirdness. The crosshair stays centered. The Milky Way and the showcase planet sweep past convincingly.

Second eight. The player taps `W`. A short thrust-on chirp plays. The engine voice rises in pitch and brightness. The ship starts to drift forward. Releasing `W` keeps the ship coasting. The HUD speed number ticks up smoothly with a small forward direction marker.

Second twelve. The player holds `W`. The acceleration curve commits — the ship clearly accelerates more aggressively the longer the key is held. After a few seconds the HUD reads several times the speed of light, the engine voice is bright and aggressive, the stars in the background are now drifting visibly.

Second twenty. The player approaches a nearby planet. The proximity speed-bleed kicks in quietly so they don't slam through. Atmosphere rim glows on the planet's silhouette. The directional light from the home star carves a clean terminator across its surface. Moons are clearly orbiting it.

Second twenty-five. The player aims at a star somewhere in the distance and clicks left mouse. The laser snaps out, a green beam to the star with a small glow at the muzzle. The aim-assist quietly catches the star inside the cone. The supernova erupts — triple flash, three shock bubbles, five colored rings, fragment cloud, ray streaks, violet sparks lingering after. The sub-bass whomp lands. The star is dead; from now on it will not appear again. The HUD's "nearest body" updates to whatever is now closest.

If all of that is true, the piece works. If any of it isn't, fix the thing that's missing before adding anything else.

---

## Polish notes

A few specific things that separate a working procedural galaxy from one that feels finished.

**Color discipline.** Pick a small palette and stay inside it. Warm star colors lean ochre and pale yellow, not saturated orange. Cool star colors are blue-white, not Crayola blue. Nebulae are the place to be saturated; everywhere else, desaturate. Atmospheres should pick up their planet's character rather than fighting it.

**Tonemapping over additive blending.** Many things in this piece are additive (stars, glows, nebulae, explosions). Without a tonemapper they will white-clip immediately. With a low-exposure ACES filmic tonemapper and a high bloom threshold, the same content reads as bright but never washes out. Treat the tonemapper and the bloom threshold as the two main brightness knobs — turn them, then re-evaluate, before adjusting individual sources.

**Layer separation.** From any vantage point the player should be able to identify three distinct depth layers — the immediate foreground (this planet, this star), the mid-distance (other planets in the system, nearby asteroid belts), and the deep background (other stars, nebulae, the Milky Way). If any two layers compete for attention they're either too close in brightness or too close in scale. Adjust until the eye knows where to land.

**Asymmetry.** Procedural content turns boring when it averages out. Don't seed every star to the median radius, don't give every planet the same number of moons, don't space asteroid belts at uniform intervals. Power-law distributions and rare outliers are what give a procedural piece soul. The "rare large boulder" inside an asteroid belt sells the entire belt.

**Negative space.** The chunk volume is mostly empty. That is correct. Don't be tempted to crank the stars-per-chunk count to fill it — empty space is what makes a star matter when you reach one. The chunk theme system already does the right thing here: most chunks have one or two stars, void chunks are nearly empty, cluster chunks are dense. Let that imbalance live.

**Transitions you don't notice.** Approaching a new star and the system materializing should feel natural — the planets fade in over a short range as the camera enters the system. Don't pop them in instantly. Tying their material opacity to a smoothstep of camera-distance-to-planet gives you a free fade with no extra logic.

**Closeup behavior.** When the camera gets within a few planet-radii of a planet, that planet has to hold up to scrutiny. The surface is vertex-colored not textured, so the vertex density of the sphere matters — go denser than you think you need on the showcase planet specifically. Cloud and atmosphere shells slightly larger than the planet sphere, with rim-Fresnel falloff, so they don't intersect awkwardly when the player flies low.

**HUD restraint.** Don't add more lines, don't add bars, don't show debug info. The two-line HUD is enough. If you find yourself wanting to add a minimap or a target list, stop — that's a different game.

---

## Depth, scale, and rendering correctness

A scene that mixes objects ranging from tens of units to hundreds of thousands of units in size, with a near-clip on the order of one tenth of a unit and a far-clip past a hundred thousand units, will z-fight in catastrophic ways under a default depth buffer. Use the logarithmic depth buffer renderer option. Then make sure every custom shader you write that participates in the world depth contains the log-depth shader includes in both vertex and fragment — without those, the custom-shaded objects will not depth-test correctly against the rest of the scene.

The atmosphere shells, cloud shells, ring discs, far-star points, near-star glow billboards, ring debris field, and skydome all need this treatment. Forgetting it on even one shader causes that layer to either disappear behind opaque geometry or punch through it inconsistently.

Sort-order discipline matters too. The black hole's black core sprite must render after its accretion disc so the disc never bleeds through the hole. The skydome must render first with depth-write disabled so it sits behind everything and contributes no depth. Transparent ring discs, cloud shells, and atmosphere shells must not write depth so they can compose correctly with one another.

For very large objects with bounding spheres Three.js can't trust (the skydome, the chunk-streamed star points, the instanced star glow billboards, instanced asteroid fields), disable frustum culling explicitly — otherwise they will pop out of view at unexpected angles.

---

## Performance and the streaming budget

The chunk-streaming system is the single most performance-sensitive piece of this file. Get its budget right.

Pick a chunk size large enough that even at top speed the player crosses a chunk boundary at most a few times per second — measured in chunk-side-lengths-per-second this is "ten or so". Pick a chunk radius (in chunks loaded around the camera) large enough that the farthest loaded star is past the back-skydome's effective viewing horizon — otherwise a stretch of newly loaded stars will appear in a halo on the horizon and read as a wall. Pick stars-per-chunk so that the active-volume star count sits in the low thousands — enough to fill the sky, low enough that the GPU stays happy.

Rebuild the active star list only when the camera's containing chunk changes. Cache chunk contents so re-entering a chunk later is instant. Rebuild the point geometry, the instanced glow billboards, and the instanced hit spheres in one pass from the active list; never update them per frame.

For everything else that lives in the inner loop:

- Avoid heap allocations entirely in the animate loop and anything it calls. Preallocated scratch vectors and matrices at module scope; reuse them by `.copy()` and `.set()`.
- Update planet positions, atmosphere uniforms, cloud uniforms, ring shadow uniforms, ring asteroid rotations, moon positions only when a system is the active system, not for every system that exists.
- Use one instanced mesh per asteroid belt and per ring debris field; never one mesh per rock.
- Throttle the nearest-star scan to once every several frames; it does not need to run per frame.

Cap the live explosion count at a small number. Each multi-phase explosion creates dozens of meshes, materials, and points objects — letting them stack will exhaust memory in long sessions. When the cap is hit and a new explosion fires, retire the oldest one cleanly: detach from scene, dispose geometries and materials, remove from the explosions list.

---

## Tunable constants

Pull all important magic numbers up into clearly named constants at the top of the script so a reader can tune the piece without hunting. The list a senior reviewer will look for includes the chunk size, the chunk radius, stars per chunk, the maximum active star count, the visible-far horizon, the fade-band width, the speed-of-light unit, the maximum speed multiplier, mouse sensitivity, roll rate, the brake duration and time constant, the throttle curve thresholds, the maximum live explosions, the bloom strength and radius and threshold, the vignette strength, the contrast, the tonemap exposure, and the planet fade-in and fade-out distances.

A reader should be able to skim those constants and immediately understand the broad parameters of the piece without reading the rest of the file.

---

## Final notes on craft

This is a one-shot piece you might never touch again. Resist features that don't pay off in the first thirty seconds of play. There is no menu. There is no scoreboard. There is no progression system. The verbs the player has are "fly" and "destroy", and the universe gives them a reason to use both.

What a senior reviewer will look for, in priority order:

1. The opening frame is a polished cinematic shot — home star, showcase planet, Milky Way, no jank.
2. The ship feels good — smooth mouse, committed throttle curve, satisfying brake, no gimbal pop.
3. The laser is satisfying — fires reliably, hits stars even when the aim is approximate, supernova is dramatic.
4. The audio is alive — pad and engine always playing, events distinct, supernova carries weight.
5. Stars look like stars — small clean cores, controlled glow, restrained twinkle, real diffraction spikes only on the close ones.
6. Planets hold up to scrutiny — surface variation, atmosphere rim, real ring debris, varied moons.
7. Performance — smooth at 1080p on a mid-range integrated GPU, no hitches when crossing chunks or activating systems.
8. Code is readable — sectioned, constants up top, comments only where the why is non-obvious, no dead code.

If the first three of those are right and the rest are reasonable, the piece is shippable. If those three aren't right, no amount of feature work elsewhere will save it.

One last principle: this is a single file, opened directly from disk, with no server, no fetch calls, no localStorage. Everything is in memory, everything is procedural, everything resets on reload. That constraint is a feature — it means a curious reader can read the entire thing end to end and understand it. Keep the code structured so they can. The point is not to demonstrate cleverness; the point is to deliver a beautiful piece of running software that does not need explaining.

---

---

## Out of scope

- No multiplayer or networking of any kind.
- No external textures, fonts, models, or audio files.
- No menus beyond the click-to-start and click-to-resume overlay.
- No physics — collision is replaced by the gentle proximity speed-bleed described above.
- No save / load.

---

## Deliverable

Output exactly one file: `galaxy.html`, ready to open in a browser. Nothing else — no README, no commentary, no separate assets. The first line of the file is `<!DOCTYPE html>` and the last line is `</html>`. Everything in between is the working program.
