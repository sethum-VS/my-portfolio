import { 
  OrthographicCamera, 
  Scene, 
  WebGLRenderer, 
  ShaderMaterial, 
  PlaneGeometry, 
  Mesh, 
  Vector2 
} from 'three';

export class WebGLNoise {
  private camera: OrthographicCamera;
  private scene: Scene;
  private renderer: WebGLRenderer;
  private material: ShaderMaterial;
  private rafId: number = 0;
  private resizeTimeout: any;

  // Reverted throttling for film grain to a stable 24fps (faster than 14, prevents lag)
  private lastTime: number = 0;
  private readonly fps: number = 50;
  private readonly interval: number = 1000 / this.fps;

  constructor(canvas: HTMLCanvasElement) {
    this.renderer = new WebGLRenderer({ canvas, alpha: true, antialias: false });
    this.renderer.setClearColor(0x000000, 0);
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    // Cap pixel ratio at 1.0 for performance (Finding #5)
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.0));

    this.scene = new Scene();
    this.camera = new OrthographicCamera(-1, 1, 1, -1, 0, 1);

    this.material = new ShaderMaterial({
      uniforms: {
        u_time_smooth: { value: 0.0 },
        u_time_noise: { value: 0.0 },
        u_resolution: { value: new Vector2(window.innerWidth, window.innerHeight) }
      },
      vertexShader: `
        void main() {
          gl_Position = vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        uniform float u_time_smooth;
        uniform float u_time_noise;
        uniform vec2 u_resolution;

        // High frequency hash for denser particle count
        float hash(vec2 p) {
            return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453123);
        }

        // Extracts a clean mathematical 4D selector mask avoiding branching
        vec4 getMask(float state) {
            return vec4(
                step(state, 0.5),
                step(0.5, state) * step(state, 1.5),
                step(1.5, state) * step(state, 2.5),
                step(2.5, state)
            );
        }

        void main() {
            vec2 uv = gl_FragCoord.xy / u_resolution.xy;
            vec2 p = uv * 2.0 - 1.0;
            p.x *= u_resolution.x / u_resolution.y;

            // 1:1 pixel coordinates
            vec2 noiseUv = gl_FragCoord.xy; 
            float noise = hash(noiseUv + u_time_noise);
            
            // Sharpened power curve back to reduce overall bright dust presence 
            noise = pow(noise, 1.55); 

            vec2 center = vec2(-1.2, 0.6);
            float radius = 2.65; 
            
            vec2 localP = p - center;
            float angle = radians(275.0);
            mat2 rot = mat2(cos(angle), -sin(angle), 
                            sin(angle), cos(angle));
            vec2 tiltedP = rot * localP;

            // =========================================================
            // SEAMLESS CONTINUOUS ANIMATION QUEUE 
            // =========================================================
            // Decoupled from the light orbit to allow continuous, visible morphing
            float cycleDuration = 4.0; // 4 seconds per state
            float timePhase = u_time_smooth / cycleDuration;
            
            float numStates = 4.0;
            float currentState = mod(floor(timePhase), numStates);
            float nextState = mod(floor(timePhase) + 1.0, numStates);
            
            // 50% hold, 50% seamless transition directly on-screen
            float transition = fract(timePhase);
            float blend = smoothstep(0.5, 1.0, transition);

            // --- PRESET 0: Classic Calm Water ---
            // Greatly toned down structure velocities and geometric distortions
            float edge0 = sin(tiltedP.y * 6.0 - u_time_smooth * 1.5) * 0.05;
            float surf0 = sin(tiltedP.y * 10.0 - u_time_smooth * 1.2 + tiltedP.x * 4.0) * 0.04;
            float pulse0 = 0.0;
            float em0 = pow(hash(noiseUv * 0.5 - u_time_smooth * 5.0), 5.0) * 0.08;

            // --- PRESET 1: Magnetic Fluid Boiling ---
            float boilPhase = u_time_smooth * 1.2;
            float edge1 = (sin(length(tiltedP) * 15.0 - boilPhase) * cos(tiltedP.x * 10.0 + boilPhase)) * 0.06; // Scaled down violence
            float surf1 = sin(tiltedP.y * 15.0 - u_time_smooth * 2.5) * 0.02;
            float pulse1 = 0.0;
            float em1 = pow(hash(noiseUv * 0.5 - u_time_smooth * 4.0), 5.0) * 0.08;

            // --- PRESET 2: Gas Giant Atmospheric Banding ---
            float edge2 = sin(tiltedP.y * 4.0 - u_time_smooth * 1.0) * 0.02; 
            float surf2 = sin(tiltedP.x * 24.0 + u_time_smooth * 3.0) * 0.06; // Subdued gas streams
            float pulse2 = 0.0;
            float em2 = pow(hash(noiseUv * 0.5 - u_time_smooth * 5.0), 5.0) * 0.04; 

            // --- PRESET 3: Chaotic Energy Storm ---
            float edge3 = sin(tiltedP.x * 20.0 + u_time_smooth * 5.0) * 0.05; 
            float surf3 = sin(tiltedP.y * 30.0 + tiltedP.x * 30.0 - u_time_smooth * 5.0) * 0.06; 
            float pulse3 = (sin(u_time_smooth * 4.0) * 0.5 + 0.5) * 0.05; 
            float em3 = pow(hash(noiseUv * 0.2 - u_time_smooth * 10.0), 4.0) * 0.15; 

            // Extract the active hardware arrays
            vec4 edgeVector = vec4(edge0, edge1, edge2, edge3);
            vec4 surfVector = vec4(surf0, surf1, surf2, surf3);
            vec4 pulseVector = vec4(pulse0, pulse1, pulse2, pulse3);
            vec4 emVector = vec4(em0, em1, em2, em3);

            // Fetch structural masks
            vec4 maskCurrent = getMask(currentState);
            vec4 maskNext = getMask(nextState);

            // Crossfade mathematically (Now visible and seamlessly integrated)
            float finalEdgeWave = mix(dot(edgeVector, maskCurrent), dot(edgeVector, maskNext), blend);
            float finalSurfaceWave = mix(dot(surfVector, maskCurrent), dot(surfVector, maskNext), blend);
            float finalPulse = mix(dot(pulseVector, maskCurrent), dot(pulseVector, maskNext), blend);
            float finalEmission = mix(dot(emVector, maskCurrent), dot(emVector, maskNext), blend);
            // =========================================================

            // Perpetual ambient base wave to ensure the edges are continuously rippling like fluid independent of queue transitions
            // Amplified significantly so it's physically distinct
            float baseEdgeWave = sin(tiltedP.y * 5.0 - u_time_smooth * 2.0) * 0.12 + 
                                 cos(tiltedP.x * 4.0 + u_time_smooth * 1.5) * 0.08;
            
            // Perpetual surface ripple so the internal light shading visibly undulates alongside the edge
            float baseSurfaceWave = sin(tiltedP.y * 8.0 - u_time_smooth * 1.5 + tiltedP.x * 3.0) * 0.08;

            float rawDist = length(tiltedP);

            float edgeAngle = atan(tiltedP.y, tiltedP.x);
            // Extended inner smoothstep bounds so the edge glow beautifully bleeds onto the globe's interior surface
            float edgeBand = smoothstep(radius - 1.2, radius - 0.02, rawDist) *
                             (1.0 - smoothstep(radius - 0.02, radius + 0.15, rawDist));

            // Circular motion around the globe with mild evolving variation.
            float runnerPhase = edgeAngle * 11.0 - u_time_smooth * 4.1;
            float runnerA = sin(runnerPhase);
            float runnerB = sin(edgeAngle * 18.0 + u_time_smooth * 2.2 + sin(u_time_smooth * 0.7) * 0.8);
            float runnerMix = 0.30 + 0.25 * (sin(u_time_smooth * 0.8) * 0.5 + 0.5);
            float edgeRunner = mix(runnerA, runnerB, runnerMix);

            float runnerPulse = 0.82 + 0.24 * sin(u_time_smooth * 1.6 + edgeAngle * 2.0);

            // Much denser multi-layer spark field for a stronger, more noticeable edge particle trail.
            float edgeSparkA = pow(hash(vec2(edgeAngle * 44.0 - u_time_smooth * 6.0, u_time_smooth * 2.9)), 2.8) * 0.030;
            float edgeSparkB = pow(hash(vec2(edgeAngle * 71.0 + 17.0 - u_time_smooth * 8.3, u_time_smooth * 3.8)), 3.5) * 0.020;
            float edgeSparkC = pow(hash(vec2(edgeAngle * 103.0 + 29.0 - u_time_smooth * 11.5, u_time_smooth * 5.2)), 4.2) * 0.012;
            float edgeSpark = edgeSparkA + edgeSparkB + edgeSparkC;
            float edgeRunnerWave = (edgeRunner * 0.12 * runnerPulse + edgeSpark * 1.1) * edgeBand;

            float distortedDist = rawDist + finalEdgeWave + baseEdgeWave + edgeRunnerWave + (noise * 0.03);

            // FIXED: Sharpened boundary. MacOS compatibility prevents undefined smoothstep bounds. Always process lower-to-upper strictly.
            // A tighter smoothstep makes the physical edge wave much more apparent
            float edgeMask = 1.0 - smoothstep(radius - 0.07, radius, distortedDist);

            float edgeRunnerGlow = edgeBand *
                                   ((0.55 + 0.65 * abs(edgeRunner)) *
                                    (0.12 + 0.05 * sin(u_time_smooth * 1.3 + edgeAngle * 3.0)) +
                                    edgeSpark * 2.8);
            float alpha = edgeRunnerGlow * edgeMask;
            vec3 color = vec3(1.0); 

            if (distortedDist < radius + 0.2) {
                
                // Use distortedDist instead of rawDist to physically deform the 3D normal,
                // blending the edge wave inward onto the globe's surface.
                float accurateDist = clamp(distortedDist, 0.0, radius); 
                float z = sqrt(radius*radius - accurateDist*accurateDist);
                vec2 deformedP = tiltedP * (distortedDist / max(rawDist, 0.001));
                vec3 normal = normalize(vec3(deformedP, z));

                // Restructured physics loop: The light orbits faster to prevent long periods of black screen
                float orbitT = u_time_smooth * 0.7;
                vec3 lightDir = normalize(vec3(
                    sin(orbitT) * 1.5, 
                    cos(orbitT * 0.5) * 0.3, 
                    // Moon-like logic: keeps a tiny crescent illuminated even when mostly on dark side
                    cos(orbitT) * 0.8 + 0.3
                ));

                float diffuse = dot(normal, lightDir);
                float distortedDiffuse = diffuse + finalSurfaceWave + baseSurfaceWave;
                
                // Dynamic Terminator Band boundary
                float terminatorBand = smoothstep(0.0, 0.25, distortedDiffuse) * (1.0 - smoothstep(0.15, 0.7, distortedDiffuse));
                
                // Queued Core illuminating pulses
                float coreLight = smoothstep(0.2, 1.0, distortedDiffuse) * (0.04 + finalPulse);

                // Reversed boundaries
                float darkSideSparkle = 1.0 - smoothstep(-0.4, 0.0, distortedDiffuse);
                float darkSideComposite = darkSideSparkle * finalEmission;

                // Restrained structural particle density formula mapping to ease the intense brightness levels back down (scaled from 0.75 to 0.45)
                float combinedIllumination = (terminatorBand * noise * 0.45) + coreLight + darkSideComposite;
                
                float globalFade = smoothstep(-0.4, 0.5, lightDir.z);

                alpha += combinedIllumination * globalFade * edgeMask;
            }

            gl_FragColor = vec4(color, alpha);
        }
      `,
      transparent: true,
      depthWrite: false,
      depthTest: false,
    });

    const geometry = new PlaneGeometry(2, 2);
    const plane = new Mesh(geometry, this.material);
    this.scene.add(plane);

    window.addEventListener('resize', this.onWindowResize);
    this.rafId = requestAnimationFrame(this.animate);
  }

  // Implement 150ms debounce on resize (Finding #8)
  private onWindowResize = () => {
    clearTimeout(this.resizeTimeout);
    this.resizeTimeout = setTimeout(() => {
      this.renderer.setSize(window.innerWidth, window.innerHeight);
      this.material.uniforms.u_resolution.value.set(window.innerWidth, window.innerHeight);
    }, 150);
  };

  private animate = (currentTime: number) => {
    this.rafId = requestAnimationFrame(this.animate);

    // Smooth 60fps for orbital mechanics
    this.material.uniforms.u_time_smooth.value = currentTime * 0.001;

    // Partially restored 24fps throttling explicitly for non-laggy film grain
    const delta = currentTime - this.lastTime;
    if (delta > this.interval) {
      this.lastTime = currentTime - (delta % this.interval);
      this.material.uniforms.u_time_noise.value = currentTime * 0.001;
    }

    this.renderer.render(this.scene, this.camera);
  };

  public destroy() {
    cancelAnimationFrame(this.rafId);
    window.removeEventListener('resize', this.onWindowResize);
    clearTimeout(this.resizeTimeout);
    this.geometryDispose();
    this.material.dispose();
    this.renderer.dispose();
  }

  private geometryDispose() {
    this.scene.children.forEach(child => {
      if (child instanceof Mesh) {
        child.geometry.dispose();
      }
    });
  }
}
