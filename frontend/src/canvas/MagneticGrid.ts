export class MagneticGrid {
  private canvas: HTMLCanvasElement;
  private ctx: CanvasRenderingContext2D;
  private nodes: { x: number; y: number; baseX: number; baseY: number }[] = [];
  
  private mouse = { x: -1000, y: -1000, active: false };
  
  // Event handler references for cleanup
  private handleResize: () => void;
  private handleMouseMove: (e: MouseEvent) => void;
  private handleMouseLeave: () => void;
  private resizeTimeout: any;
  
  // Configuration
  private readonly gridSize = 40; // 40px blocks
  private readonly magnetRadius = 200; // Pull radius
  private readonly magnetStrength = 0.6; // How strong the pull is
  
  // Animation frame id
  private rafId: number = 0;
  private alpha: number;

  constructor(canvas: HTMLCanvasElement, alpha: number = 0.25) {
    this.canvas = canvas;
    this.alpha = alpha;
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Could not initialize 2D context");
    this.ctx = context;

    this.handleResize = () => {
      clearTimeout(this.resizeTimeout);
      this.resizeTimeout = setTimeout(() => {
        this.resize();
        this.draw(); // Ensure it redraws immediately after resize
      }, 150);
    };

    this.handleMouseMove = (e: MouseEvent) => {
      this.mouse.x = e.clientX;
      this.mouse.y = e.clientY;
      this.mouse.active = true;
      if (!this.rafId) {
        this.rafId = requestAnimationFrame(this.loop);
      }
    };

    this.handleMouseLeave = () => {
      this.mouse.active = false;
    };

    this.init();
    this.bindEvents();
    
    // Draw initial state
    this.draw();
    // Start loop in case nodes need to settle, but it will sleep if inactive
    this.rafId = requestAnimationFrame(this.loop);
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
    window.addEventListener("resize", this.handleResize);
    window.addEventListener("mousemove", this.handleMouseMove, { passive: true });
    window.addEventListener("mouseleave", this.handleMouseLeave);
  }

  private update(): boolean {
    let needsUpdate = false;
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
      
      const moveX = (targetX - node.x) * 0.15;
      const moveY = (targetY - node.y) * 0.15;
      
      node.x += moveX;
      node.y += moveY;
      
      if (Math.abs(moveX) > 0.01 || Math.abs(moveY) > 0.01) {
        needsUpdate = true;
      }
    }
    return needsUpdate;
  }

  private draw() {
    this.ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    
    for (const node of this.nodes) {
      this.ctx.beginPath();
      const displacementX = node.x - node.baseX;
      const displacementY = node.y - node.baseY;
      const displacement = Math.sqrt(displacementX * displacementX + displacementY * displacementY);
      
      // Fixed aesthetic, purely positional magnetic displacement
      this.ctx.fillStyle = `rgba(88, 199, 255, ${this.alpha})`;
      this.ctx.arc(node.x, node.y, 1.5, 0, Math.PI * 2);
      this.ctx.fill();
    }
  }

  private loop = () => {
    const needsUpdate = this.update();
    this.draw();
    
    if (this.mouse.active || needsUpdate) {
      this.rafId = requestAnimationFrame(this.loop);
    } else {
      this.rafId = 0; // Sleep
    }
  };
  
  public destroy() {
    if (this.rafId) {
      cancelAnimationFrame(this.rafId);
      this.rafId = 0;
    }
    window.removeEventListener("resize", this.handleResize);
    window.removeEventListener("mousemove", this.handleMouseMove);
    window.removeEventListener("mouseleave", this.handleMouseLeave);
  }
}
