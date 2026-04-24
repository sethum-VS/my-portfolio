import { MagneticGrid } from "./canvas/MagneticGrid";
import { WebGLNoise } from "./canvas/WebGLNoise";

let gridInstances: MagneticGrid[] = [];
let noiseInstance: WebGLNoise | null = null;

function initCanvas() {
  const baseCanvas = document.getElementById("magnetic-canvas-base") as HTMLCanvasElement;
  const waveCanvas = document.getElementById("magnetic-canvas-wave") as HTMLCanvasElement;
  const noiseCanvas = document.getElementById("webgl-noise-canvas") as HTMLCanvasElement;
  
  if (baseCanvas && gridInstances.length === 0) {
    gridInstances.push(new MagneticGrid(baseCanvas, 0.25));
  }
  if (waveCanvas && gridInstances.length === 1) {
    gridInstances.push(new MagneticGrid(waveCanvas, 0.9));
  }
  if (noiseCanvas && !noiseInstance) {
    noiseInstance = new WebGLNoise(noiseCanvas);
  }
}

function initCursor() {
  const cursor = document.getElementById("custom-cursor");
  if (!cursor) return;

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
    cursor!.style.transform = `translate(calc(${targetX}px - 50%), calc(${targetY}px - 50%))`;
    requestAnimationFrame(renderCursor);
  }
  requestAnimationFrame(renderCursor);
}

function initNavBlob() {
  const blob = document.getElementById("nav-blob");
  const container = document.getElementById("desktop-nav-links");
  if (!blob || !container) return;

  function moveBlob(target: HTMLElement) {
    const targetRect = target.getBoundingClientRect();
    const containerRect = container!.getBoundingClientRect();
    blob!.style.width = `${targetRect.width}px`;
    blob!.style.height = `${targetRect.height}px`;
    blob!.style.left = `${targetRect.left - containerRect.left}px`;
    blob!.style.top = `${targetRect.top - containerRect.top}px`;
  }

  // Initial position
  const activeLink = container.querySelector('[data-active="true"]') as HTMLElement;
  if (activeLink) {
    blob.style.transition = 'none';
    moveBlob(activeLink);
    blob.offsetHeight; // force reflow
    blob.style.transition = '';
    blob.style.opacity = "1";
  }

  // Handle clicks for animation and exit transition
  const links = container.querySelectorAll('.nav-link');
  links.forEach(link => {
    link.addEventListener('click', (e) => {
      links.forEach(l => {
        l.classList.remove('font-bold', 'text-white');
        l.classList.add('font-semibold', 'text-zinc-400');
        l.setAttribute('data-active', 'false');
      });
      link.classList.add('font-bold', 'text-white');
      link.classList.remove('font-semibold', 'text-zinc-400');
      link.setAttribute('data-active', 'true');
      
      moveBlob(link as HTMLElement);

      // Trigger SPA page exit transition on main-content immediately
      const mainContent = document.getElementById('main-content');
      if (mainContent) {
        mainContent.classList.add('page-transition-exit');
      }
    });
  });

  // Handle browser back/forward buttons
  window.addEventListener('popstate', () => {
    const path = window.location.pathname;
    links.forEach(link => {
      if (link.getAttribute('href') === path) {
        link.classList.add('font-bold', 'text-white');
        link.classList.remove('font-semibold', 'text-zinc-400');
        link.setAttribute('data-active', 'true');
        moveBlob(link as HTMLElement);
      } else {
        link.classList.remove('font-bold', 'text-white');
        link.classList.add('font-semibold', 'text-zinc-400');
        link.setAttribute('data-active', 'false');
      }
    });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  initCanvas();
  initCursor();
  initNavBlob();
});

document.addEventListener("htmx:afterSwap", () => {
  gridInstances.forEach(g => g.destroy());
  gridInstances = [];
  if (noiseInstance) {
    noiseInstance.destroy();
    noiseInstance = null;
  }
  initCanvas();
  initCursor();
});
