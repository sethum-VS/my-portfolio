(function () {
  const ACTION = "resume_request";
  const RECAPTCHA_W = 304;
  const RECAPTCHA_H = 78;
  let widgetId = null;
  let scriptLoading = false;
  let resizeObserver = null;

  function setToken(token) {
    var el = document.getElementById("g-recaptcha-response");
    if (el) el.value = token || "";
  }

  function siteKeyFromDOM() {
    var widget = document.getElementById("recaptcha-widget");
    if (widget && widget.dataset.sitekey) return widget.dataset.sitekey;
    var form = document.querySelector(
      '#resume-modal-inner form[data-recaptcha-sitekey]'
    );
    return form ? form.getAttribute("data-recaptcha-sitekey") : "";
  }

  function showFormError(message) {
    var err = document.getElementById("resume-form-error");
    if (!err) return;
    err.textContent = message;
    err.classList.remove("hidden");
  }

  function clearFormError() {
    var err = document.getElementById("resume-form-error");
    if (!err) return;
    err.textContent = "";
    err.classList.add("hidden");
  }

  function loadEnterpriseScript() {
    return new Promise(function (resolve, reject) {
      if (window.grecaptcha && window.grecaptcha.enterprise) {
        resolve();
        return;
      }
      if (scriptLoading) {
        document.addEventListener(
          "recaptcha-enterprise-loaded",
          function () {
            resolve();
          },
          { once: true }
        );
        return;
      }
      scriptLoading = true;
      window.__resumeRecaptchaOnload = function () {
        document.dispatchEvent(new Event("recaptcha-enterprise-loaded"));
        resolve();
      };
      var s = document.createElement("script");
      s.src =
        "https://www.google.com/recaptcha/enterprise.js?onload=__resumeRecaptchaOnload&render=explicit";
      s.async = true;
      s.defer = true;
      s.setAttribute("data-resume-recaptcha", "true");
      s.onerror = function () {
        reject(new Error("Failed to load reCAPTCHA Enterprise"));
      };
      document.head.appendChild(s);
    });
  }

  function recaptchaWrap() {
    return document.querySelector(".recaptcha-responsive");
  }

  function recaptchaInner() {
    var widget = document.getElementById("recaptcha-widget");
    return widget ? widget.firstElementChild : null;
  }

  function disconnectRecaptchaFit() {
    if (resizeObserver) {
      resizeObserver.disconnect();
      resizeObserver = null;
    }
    var wrap = recaptchaWrap();
    if (wrap) wrap.style.height = "";
    var inner = recaptchaInner();
    if (inner) inner.style.transform = "";
  }

  function fitRecaptcha() {
    var wrap = recaptchaWrap();
    var inner = recaptchaInner();
    if (!wrap || !inner) return;

    var available = wrap.getBoundingClientRect().width;
    if (available <= 0) return;

    var scale = Math.min(1, available / RECAPTCHA_W);
    inner.style.transformOrigin = "top left";
    inner.style.transform = scale < 1 ? "scale(" + scale + ")" : "";
    wrap.style.height = RECAPTCHA_H * scale + "px";
  }

  function observeRecaptchaFit() {
    disconnectRecaptchaFit();
    var wrap = recaptchaWrap();
    if (!wrap || typeof ResizeObserver === "undefined") return;

    resizeObserver = new ResizeObserver(function () {
      fitRecaptcha();
    });
    resizeObserver.observe(wrap);
    if (wrap.parentElement) resizeObserver.observe(wrap.parentElement);
  }

  function scheduleRecaptchaFit() {
    fitRecaptcha();
    var attempts = 0;
    function tick() {
      fitRecaptcha();
      if (!recaptchaInner()) {
        if (++attempts < 60) requestAnimationFrame(tick);
        return;
      }
      observeRecaptchaFit();
      fitRecaptcha();
    }
    requestAnimationFrame(tick);
  }

  function resetWidget() {
    disconnectRecaptchaFit();
    if (
      widgetId !== null &&
      window.grecaptcha &&
      window.grecaptcha.enterprise
    ) {
      try {
        window.grecaptcha.enterprise.reset(widgetId);
      } catch (_) {}
    }
    widgetId = null;
    setToken("");
    var container = document.getElementById("recaptcha-widget");
    if (container) container.innerHTML = "";
  }

  function syncTokenFromWidget() {
    if (
      widgetId === null ||
      !window.grecaptcha ||
      !window.grecaptcha.enterprise
    ) {
      return;
    }
    try {
      var token = window.grecaptcha.enterprise.getResponse(widgetId);
      if (token) setToken(token);
    } catch (_) {}
  }

  function renderWidget() {
    var container = document.getElementById("recaptcha-widget");
    if (!container || widgetId !== null) return;

    var siteKey = siteKeyFromDOM();
    if (!siteKey) return;

    loadEnterpriseScript()
      .then(function () {
        if (!document.getElementById("recaptcha-widget")) return;
        if (widgetId !== null) return;

        widgetId = window.grecaptcha.enterprise.render(container, {
          sitekey: siteKey,
          action: ACTION,
          theme: "dark",
          callback: function (token) {
            setToken(token);
            clearFormError();
          },
          "expired-callback": function () {
            setToken("");
          },
          "error-callback": function () {
            setToken("");
          },
        });
        scheduleRecaptchaFit();
      })
      .catch(function (err) {
        console.error("reCAPTCHA render failed:", err);
        showFormError("Could not load verification. Refresh and try again.");
      });
  }

  function onModalSwap(evt) {
    var target = evt.detail && evt.detail.target;
    if (!target) return;
    if (
      target.id === "resume-modal-root" ||
      target.id === "resume-modal-inner"
    ) {
      resetWidget();
      clearFormError();
      renderWidget();
    }
  }

  document.body.addEventListener("htmx:afterSwap", onModalSwap);
  document.body.addEventListener("resume-modal-closed", resetWidget);

  var resizeTimer;
  window.addEventListener("resize", function () {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(fitRecaptcha, 50);
  });
  window.addEventListener("orientationchange", function () {
    setTimeout(fitRecaptcha, 150);
  });

  document.body.addEventListener("htmx:beforeRequest", function (evt) {
    var elt = evt.detail && evt.detail.elt;
    if (!elt || elt.getAttribute("hx-post") !== "/api/resume/request") return;

    if (!siteKeyFromDOM()) return;

    syncTokenFromWidget();
    var tokenEl = document.getElementById("g-recaptcha-response");
    if (!tokenEl || !tokenEl.value) {
      evt.preventDefault();
      showFormError("Complete the reCAPTCHA challenge before submitting.");
    }
  });
})();
