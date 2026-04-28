import { MagneticGrid } from "./canvas/MagneticGrid";
import { WebGLNoise } from "./canvas/WebGLNoise";

let gridInstances: MagneticGrid[] = [];
let noiseInstance: WebGLNoise | null = null;
let contentProtectionInitialized = false;

function initContentProtection() {
  if (contentProtectionInitialized) return;
  contentProtectionInitialized = true;

  const block = (event: Event) => {
    event.preventDefault();
  };

  const blockCopyShortcuts = (event: KeyboardEvent) => {
    const key = event.key.toLowerCase();
    const hasModifier = event.ctrlKey || event.metaKey;

    // Block common copy and select-all shortcuts, including Shift+Insert paste-copy pathways.
    const blockedCombination =
      (hasModifier && (key === "c" || key === "x" || key === "a" || key === "insert")) ||
      (event.shiftKey && key === "insert") ||
      key === "contextmenu" ||
      (event.shiftKey && key === "f10");

    if (blockedCombination) {
      event.preventDefault();
      event.stopPropagation();
    }
  };

  document.addEventListener("contextmenu", block);
  document.addEventListener("copy", block);
  document.addEventListener("cut", block);
  document.addEventListener("selectstart", block);
  document.addEventListener("dragstart", block);
  document.addEventListener("keydown", blockCopyShortcuts);
}

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
  // Skip cursor initialization entirely on touch devices
  if (window.matchMedia("(pointer: coarse)").matches) return;

  const cursor = document.getElementById("custom-cursor");
  if (!cursor) return;

  let targetX = -100;
  let targetY = -100;
  let isMoving = false;
  let cursorRafId: number | null = null;

  window.addEventListener("mousemove", (e) => {
    targetX = e.clientX;
    targetY = e.clientY;
    if (cursor.style.opacity !== "1") {
      cursor.style.opacity = "1";
    }
    
    if (!isMoving) {
      isMoving = true;
      if (cursorRafId === null) {
        cursorRafId = requestAnimationFrame(renderCursor);
      }
    }
    
    // reset moving state slightly later to pause rendering when static
    clearTimeout((window as any).cursorTimeout);
    (window as any).cursorTimeout = setTimeout(() => {
      isMoving = false;
    }, 100);
  }, { passive: true });

  function renderCursor() {
    cursor!.style.transform = `translate(calc(${targetX}px - 50%), calc(${targetY}px - 50%))`;
    
    if (isMoving) {
      cursorRafId = requestAnimationFrame(renderCursor);
    } else {
      cursorRafId = null; // stop looping when idle
    }
  }
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

  // Update mobile nav active states
  const mobileContainer = document.getElementById("mobile-nav");
  if (mobileContainer) {
    const mobileLinks = mobileContainer.querySelectorAll('a[href^="/"]');
    for (const link of Array.from(mobileLinks)) {
      const el = link as HTMLElement;
      const href = el.getAttribute('href');
      if (!href) continue;
      
      const isHome = href === '/' || href === '/home';
      const isPathHome = path === '/' || path === '/home';
      
      const isMatch = (isHome && isPathHome) || 
                     (!isHome && href !== '/' && path.startsWith(href));
                     
      if (isMatch) {
        el.classList.add('text-[#abdeff]');
        el.classList.remove('text-[#f0f0f0]', 'opacity-50', 'hover:text-[#ffcdbd]', 'hover:opacity-100');
      } else {
        el.classList.remove('text-[#abdeff]');
        el.classList.add('text-[#f0f0f0]', 'opacity-50', 'hover:text-[#ffcdbd]', 'hover:opacity-100');
      }
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

/**
 * Toggles the mobile navigation overlay with opening/closing animations.
 * @param isOpen - true to open, false to close
 */
function toggleMobileNav(isOpen: boolean) {
  const nav = document.getElementById('mobile-nav');
  if (!nav) return;

  if (isOpen) {
    nav.classList.remove('hidden', 'closing');
    nav.classList.add('flex');
  } else {
    nav.classList.add('closing');
    
    // Duration matches the fade-out animation.
    setTimeout(() => {
      if (nav.classList.contains('closing')) {
        nav.classList.add('hidden');
        nav.classList.remove('closing', 'flex');
      }
    }, 400);
  }
}

(window as any).toggleMobileNav = toggleMobileNav;

// ── Global Event Listeners ──────────────────────────────────────────────────

document.addEventListener("DOMContentLoaded", () => {
  initContentProtection();
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
