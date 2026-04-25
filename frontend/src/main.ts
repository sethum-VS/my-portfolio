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

function destroyCanvas() {
  gridInstances.forEach(g => g.destroy());
  gridInstances = [];
  if (noiseInstance) {
    noiseInstance.destroy();
    noiseInstance = null;
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

/**
 * Updates the navigation active state and positions the blob based on the current URL path.
 * This handles cases where navigation is triggered by elements outside the navbar (like body buttons).
 */
function updateNavActiveState(animate: boolean = true) {
  const path = window.location.pathname;
  const container = document.getElementById("desktop-nav-links");
  const blob = document.getElementById("nav-blob");
  if (!container || !blob) return;

  const links = container.querySelectorAll('.nav-link');
  let activeLink: HTMLElement | null = null;

  for (const link of Array.from(links)) {
    const el = link as HTMLElement;
    const href = el.getAttribute('href');
    
    // Normalization: treat / and /home as the same for the Home link
    const isHome = href === '/' || href === '/home';
    const isPathHome = path === '/' || path === '/home';
    
    const isMatch = (isHome && isPathHome) || 
                   (!isHome && href !== '/' && path.startsWith(href!));
    
    if (isMatch) {
      el.classList.add('font-bold', 'text-white');
      el.classList.remove('font-semibold', 'text-zinc-400');
      el.setAttribute('data-active', 'true');
      activeLink = el;
    } else {
      el.classList.remove('font-bold', 'text-white');
      el.classList.add('font-semibold', 'text-zinc-400');
      el.setAttribute('data-active', 'false');
    }
  }

  if (activeLink) {
    moveBlobTo(activeLink, animate);
  } else {
    blob.style.opacity = "0";
  }
}

function moveBlobTo(target: HTMLElement, animate: boolean = true) {
  const blob = document.getElementById("nav-blob");
  const container = document.getElementById("desktop-nav-links");
  if (!blob || !container) return;

  const targetRect = target.getBoundingClientRect();
  const containerRect = container.getBoundingClientRect();
  
  if (!animate) {
    blob.style.transition = 'none';
  } else {
    blob.style.transition = ''; // Restore default CSS transition
  }

  blob.style.width = `${targetRect.width}px`;
  blob.style.height = `${targetRect.height}px`;
  blob.style.left = `${targetRect.left - containerRect.left}px`;
  blob.style.top = `${targetRect.top - containerRect.top}px`;
  blob.style.opacity = "1";

  if (!animate) {
    blob.offsetHeight; // force reflow
    blob.style.transition = '';
  }
}

function initNavBlob() {
  const container = document.getElementById("desktop-nav-links");
  if (!container) return;

  // Initial update (snap to position)
  updateNavActiveState(false);

  // Handle clicks for instant feedback and SPA exit transitions
  const links = container.querySelectorAll('.nav-link');
  links.forEach(link => {
    link.addEventListener('click', () => {
      const el = link as HTMLElement;
      
      // Visual feedback: Move blob immediately on click
      moveBlobTo(el, true);
      
      // Set clicked link to active state visually
      links.forEach(l => {
        l.classList.remove('font-bold', 'text-white');
        l.classList.add('font-semibold', 'text-zinc-400');
        l.setAttribute('data-active', 'false');
      });
      el.classList.add('font-bold', 'text-white');
      el.classList.remove('font-semibold', 'text-zinc-400');
      el.setAttribute('data-active', 'true');

      // Trigger SPA page exit transition on main-content immediately
      const mainContent = document.getElementById('main-content');
      if (mainContent) {
        mainContent.classList.add('page-transition-exit');
      }
    });
  });
}

/**
 * Strips the exit animation class from #main-content so HTMX doesn't
 * cache invisible content into its history snapshot.
 */
function cleanExitAnimation() {
  const mainContent = document.getElementById('main-content');
  if (mainContent) {
    mainContent.classList.remove('page-transition-exit');
  }
}

/**
 * Full reinitialization after page content changes (swap or history restore).
 */
function reinitPage() {
  cleanExitAnimation();
  destroyCanvas();
  initCanvas();
  initCursor();

  const blob = document.getElementById("nav-blob");
  if (blob) {
    updateNavActiveState(true);
  } else {
    initNavBlob();
  }
}

// ── Global Event Listeners ──────────────────────────────────────────────────

document.addEventListener("DOMContentLoaded", () => {
  initCanvas();
  initCursor();
  initNavBlob();
});

// Before HTMX snapshots the page into its history cache, strip the exit
// animation so the cached version has full-opacity content.
document.addEventListener("htmx:beforeHistorySave", () => {
  cleanExitAnimation();
});

// Fires after a normal HTMX AJAX swap (forward navigation).
document.addEventListener("htmx:afterSwap", () => {
  reinitPage();
});

// Fires when HTMX restores a page from its history cache (back/forward button).
// This does NOT fire htmx:afterSwap, so we need a separate handler.
document.addEventListener("htmx:historyRestore", () => {
  reinitPage();
});
