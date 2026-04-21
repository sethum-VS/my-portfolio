export class MagneticGrid {
  private canvas: HTMLCanvasElement;
  private ctx: CanvasRenderingContext2D;
  private nodes: { x: number; y: number; baseX: number; baseY: number }[] = [];
  
  private mouse = { x: -1000, y: -1000, active: false };
  
  // Configuration
  private readonly gridSize = 40; // 40px blocks
  private readonly magnetRadius = 200; // Pull radius
  private readonly magnetStrength = 0.6; // How strong the pull is
  
  // Animation frame id
  private rafId: number = 0;

  constructor(canvas: HTMLCanvasElement) {
    this.canvas = canvas;
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Could not initialize 2D context");
    this.ctx = context;

    this.init();
    this.bindEvents();
    this.loop();
  }

  private init() {
    this.resize();
  }

  private resize() {
    // To support high-DPI displays (retina)
    const dpr = window.devicePixelRatio || 1;
    this.canvas.width = window.innerWidth * dpr;
    this.canvas.height = window.innerHeight * dpr;
    
    // The CSS size
    this.canvas.style.width = `${window.innerWidth}px`;
    this.canvas.style.height = `${window.innerHeight}px`;
    
    this.ctx.scale(dpr, dpr);
    
    this.buildGrid();
  }

  private buildGrid() {
    this.nodes = [];
    
    const cols = Math.ceil(window.innerWidth / this.gridSize) + 1;
    const rows = Math.ceil(window.innerHeight / this.gridSize) + 1;
    
    // Add margin for seamless pull from edges
    for (let r = -1; r <= rows; r++) {
      for (let c = -1; c <= cols; c++) {
        const x = c * this.gridSize;
        const y = r * this.gridSize;
        this.nodes.push({ x, y, baseX: x, baseY: y });
      }
    }
  }

  private bindEvents() {
    window.addEventListener("resize", () => {
      this.resize();
    });

    window.addEventListener("mousemove", (e) => {
      this.mouse.x = e.clientX;
      this.mouse.y = e.clientY;
      this.mouse.active = true;
    });

    window.addEventListener("mouseleave", () => {
      this.mouse.active = false;
    });
  }

  private update() {
    // Loop through each node to apply spring physics and magnetism
    for (const node of this.nodes) {
      let dx = 0;
      let dy = 0;
      
      // If mouse is active, apply magnetic pull
      if (this.mouse.active) {
        const distX = this.mouse.x - node.baseX;
        const distY = this.mouse.y - node.baseY;
        const dist = Math.sqrt(distX * distX + distY * distY);
        
        if (dist < this.magnetRadius) {
          // Attract upwards towards the mouse
          const pull = 1 - dist / this.magnetRadius; // 0 to 1
          dx = distX * pull * this.magnetStrength;
          dy = distY * pull * this.magnetStrength;
        }
      }
      
      // Simple spring return-to-base interpolation
      const targetX = node.baseX + dx;
      const targetY = node.baseY + dy;
      
      node.x += (targetX - node.x) * 0.15; // Smooth spring
      node.y += (targetY - node.y) * 0.15;
    }
  }

  private draw() {
    this.ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    
    // Draw only nodes (no connecting lines since the requirement focused on node attraction)
    this.ctx.fillStyle = "#58c7ff"; // primary-container color
    
    for (const node of this.nodes) {
      this.ctx.beginPath();
      // Draw standard size dots - increasing radius slightly if pulled
      const displacementX = node.x - node.baseX;
      const displacementY = node.y - node.baseY;
      const displacement = Math.sqrt(displacementX*displacementX + displacementY*displacementY);
      
      const radius = 1.0 + (displacement * 0.05);
      
      this.ctx.arc(node.x, node.y, Math.min(radius, 3.5), 0, Math.PI * 2);
      this.ctx.fill();
    }
  }

  private loop = () => {
    this.update();
    this.draw();
    this.rafId = requestAnimationFrame(this.loop);
  };
  
  public destroy() {
    cancelAnimationFrame(this.rafId);
  }
}
