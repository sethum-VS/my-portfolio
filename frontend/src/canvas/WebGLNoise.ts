import * as THREE from 'three';

export class WebGLNoise {
  private camera: THREE.OrthographicCamera;
  private scene: THREE.Scene;
  private renderer: THREE.WebGLRenderer;
  private material: THREE.ShaderMaterial;
  private rafId: number = 0;

  // Reverted throttling for film grain to a stable 24fps (faster than 14, prevents lag)
  private lastTime: number = 0;
  private readonly fps: number = 50;
  private readonly interval: number = 1000 / this.fps;

  constructor(canvas: HTMLCanvasElement) {
    this.renderer = new THREE.WebGLRenderer({ canvas, alpha: true, antialias: false });
    this.renderer.setClearColor(0x000000, 0);
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setPixelRatio(window.devicePixelRatio);

    this.scene = new THREE.Scene();
    this.camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);

    this.material = new THREE.ShaderMaterial({
      uniforms: {
        u_time_smooth: { value: 0.0 },
        u_time_noise: { value: 0.0 },
        u_resolution: { value: new THREE.Vector2(window.innerWidth, window.innerHeight) }
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

        void main() {
            vec2 uv = gl_FragCoord.xy / u_resolution.xy;
            vec2 p = uv * 2.0 - 1.0;
            p.x *= u_resolution.x / u_resolution.y;

            // Strict 1:1 pixel coordinates for absolute minimum physical particle size
            vec2 noiseUv = gl_FragCoord.xy; 
            float noise = hash(noiseUv + u_time_noise);
            
            // Lowering the power curve exponentially increases the active amount/density of particles hitting the screen globally
            noise = pow(noise, 1.25);

            // Globe explicitly anchored 
            vec2 center = vec2(-1.2, 0.6);
            float radius = 3.0; 
            
            vec2 localP = p - center;
            
            float angle = radians(275.0);
            mat2 rot = mat2(cos(angle), -sin(angle), 
                            sin(angle), cos(angle));
            vec2 tiltedP = rot * localP;

            // 2X Speed Increase on edge wave animations
            float edgeWave = sin(tiltedP.y * 6.0 - u_time_smooth * 3.0) * 0.15;
            float rawDist = length(tiltedP);
            float distortedDist = rawDist + edgeWave + (noise * 0.04);

            float edgeMask = smoothstep(radius, radius - 0.5, distortedDist);

            float alpha = 0.0;
            vec3 color = vec3(1.0); 

            if (distortedDist < radius + 0.2) {
                
                float accurateDist = clamp(rawDist, 0.0, radius); 
                float z = sqrt(radius*radius - accurateDist*accurateDist);
                vec3 normal = normalize(vec3(tiltedP, z));

                // 2X Speed Increase on globe light translation orbit
                float t = u_time_smooth * 0.5;
                vec3 lightDir = normalize(vec3(
                    sin(t) * 1.5, 
                    cos(t * 0.6) * 0.4, 
                    sin(t * 0.4) * 0.6 + 0.6 
                ));

                float diffuse = dot(normal, lightDir);
                
                // 2X Speed Increase on interior surface waves
                float surfaceWave = sin(tiltedP.y * 10.0 - u_time_smooth * 2.4 + tiltedP.x * 4.0) * 0.08;
                float distortedDiffuse = diffuse + surfaceWave;
                
                // 1. The Terminator Band (narrowed and smoothed for less aggression)
                float terminatorBand = smoothstep(0.0, 0.25, distortedDiffuse) * (1.0 - smoothstep(0.15, 0.7, distortedDiffuse));
                
                // 2. Soft illuminated core
                float coreLight = smoothstep(0.2, 1.0, distortedDiffuse) * 0.04;

                // 3. Small sparkling animations actively flashing in the dark side
                float darkSideSparkle = smoothstep(0.0, -0.4, distortedDiffuse);
                float darkNoise = hash(noiseUv * 0.5 - u_time_smooth * 10.0);
                float darkSideEmission = darkSideSparkle * pow(darkNoise, 5.0) * 0.1;

                // Particle count doubled organically via noise mapping calculation above
                float combinedIllumination = (terminatorBand * noise * 0.5) + coreLight + darkSideEmission;
                
                float globalFade = smoothstep(-0.4, 0.5, lightDir.z);

                alpha = combinedIllumination * globalFade * edgeMask;
            }

            gl_FragColor = vec4(color, alpha);
        }
      `,
      transparent: true,
      depthWrite: false,
      depthTest: false,
    });

    const geometry = new THREE.PlaneGeometry(2, 2);
    const plane = new THREE.Mesh(geometry, this.material);
    this.scene.add(plane);

    window.addEventListener('resize', this.onWindowResize);
    this.rafId = requestAnimationFrame(this.animate);
  }

  private onWindowResize = () => {
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.material.uniforms.u_resolution.value.set(window.innerWidth, window.innerHeight);
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
    this.geometryDispose();
    this.material.dispose();
    this.renderer.dispose();
  }

  private geometryDispose() {
    this.scene.children.forEach(child => {
      if (child instanceof THREE.Mesh) {
        child.geometry.dispose();
      }
    });
  }
}
