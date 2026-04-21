(() => {
  // frontend/src/canvas/MagneticGrid.ts
  var MagneticGrid = class {
    canvas;
    ctx;
    nodes = [];
    mouse = { x: -1e3, y: -1e3, active: false };
    // Configuration
    gridSize = 40;
    // 40px blocks
    magnetRadius = 200;
    // Pull radius
    magnetStrength = 0.6;
    // How strong the pull is
    // Animation frame id
    rafId = 0;
    alpha;
    constructor(canvas, alpha = 0.25) {
      this.canvas = canvas;
      this.alpha = alpha;
      const context = canvas.getContext("2d");
      if (!context)
        throw new Error("Could not initialize 2D context");
      this.ctx = context;
      this.init();
      this.bindEvents();
      this.loop();
    }
    init() {
      this.resize();
    }
    resize() {
      const dpr = window.devicePixelRatio || 1;
      this.canvas.width = window.innerWidth * dpr;
      this.canvas.height = window.innerHeight * dpr;
      this.canvas.style.width = `${window.innerWidth}px`;
      this.canvas.style.height = `${window.innerHeight}px`;
      this.ctx.scale(dpr, dpr);
      this.buildGrid();
    }
    buildGrid() {
      this.nodes = [];
      const cols = Math.ceil(window.innerWidth / this.gridSize) + 1;
      const rows = Math.ceil(window.innerHeight / this.gridSize) + 1;
      for (let r = -1; r <= rows; r++) {
        for (let c = -1; c <= cols; c++) {
          const x = c * this.gridSize;
          const y = r * this.gridSize;
          this.nodes.push({ x, y, baseX: x, baseY: y });
        }
      }
    }
    bindEvents() {
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
    update() {
      for (const node of this.nodes) {
        let dx = 0;
        let dy = 0;
        if (this.mouse.active) {
          const distX = this.mouse.x - node.baseX;
          const distY = this.mouse.y - node.baseY;
          const dist = Math.sqrt(distX * distX + distY * distY);
          if (dist < this.magnetRadius) {
            const pull = 1 - dist / this.magnetRadius;
            dx = distX * pull * this.magnetStrength;
            dy = distY * pull * this.magnetStrength;
          }
        }
        const targetX = node.baseX + dx;
        const targetY = node.baseY + dy;
        node.x += (targetX - node.x) * 0.15;
        node.y += (targetY - node.y) * 0.15;
      }
    }
    draw() {
      this.ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
      for (const node of this.nodes) {
        this.ctx.beginPath();
        const displacementX = node.x - node.baseX;
        const displacementY = node.y - node.baseY;
        const displacement = Math.sqrt(displacementX * displacementX + displacementY * displacementY);
        this.ctx.fillStyle = `rgba(88, 199, 255, ${this.alpha})`;
        this.ctx.arc(node.x, node.y, 1.5, 0, Math.PI * 2);
        this.ctx.fill();
      }
    }
    loop = () => {
      this.update();
      this.draw();
      this.rafId = requestAnimationFrame(this.loop);
    };
    destroy() {
      cancelAnimationFrame(this.rafId);
    }
  };

  // frontend/src/main.ts
  var gridInstances = [];
  function initCanvas() {
    const baseCanvas = document.getElementById("magnetic-canvas-base");
    const waveCanvas = document.getElementById("magnetic-canvas-wave");
    if (baseCanvas && gridInstances.length === 0) {
      gridInstances.push(new MagneticGrid(baseCanvas, 0.25));
    }
    if (waveCanvas && gridInstances.length === 1) {
      gridInstances.push(new MagneticGrid(waveCanvas, 0.9));
    }
  }
  function initCursor() {
    const cursor = document.getElementById("custom-cursor");
    if (!cursor)
      return;
    let targetX = -100;
    let targetY = -100;
    window.addEventListener("mousemove", (e) => {
      targetX = e.clientX;
      targetY = e.clientY;
      if (cursor.style.opacity !== "1") {
        cursor.style.opacity = "1";
      }
    }, { passive: true });
    function renderCursor() {
      cursor.style.transform = `translate(calc(${targetX}px - 50%), calc(${targetY}px - 50%))`;
      requestAnimationFrame(renderCursor);
    }
    requestAnimationFrame(renderCursor);
  }
  document.addEventListener("DOMContentLoaded", () => {
    initCanvas();
    initCursor();
  });
  document.addEventListener("htmx:afterSwap", () => {
    gridInstances.forEach((g) => g.destroy());
    gridInstances = [];
    initCanvas();
    initCursor();
  });
})();
