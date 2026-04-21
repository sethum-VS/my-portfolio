import { MagneticGrid } from "./canvas/MagneticGrid";

let gridInstance: MagneticGrid | null = null;

function initCanvas() {
  const canvas = document.getElementById("magnetic-canvas") as HTMLCanvasElement;
  if (canvas && !gridInstance) {
    gridInstance = new MagneticGrid(canvas);
  }
}

// Full page load (direct navigation to /home)
document.addEventListener("DOMContentLoaded", initCanvas);

// HTMX swap (splash → home transition via outerHTML swap)
// After HTMX swaps content, DOMContentLoaded won't fire again,
// so we listen for htmx:afterSwap to re-initialize the canvas.
document.addEventListener("htmx:afterSwap", () => {
  // Destroy previous instance if any
  if (gridInstance) {
    gridInstance.destroy();
    gridInstance = null;
  }
  initCanvas();
});
