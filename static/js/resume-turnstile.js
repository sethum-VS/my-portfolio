(function () {
  const ACTION = "resume_request";
  let widgetId = null;
  let scriptLoading = false;

  function setToken(token) {
    var el = document.getElementById("cf-turnstile-response");
    if (el) el.value = token || "";
  }

  function siteKeyFromDOM() {
    var widget = document.getElementById("turnstile-widget");
    if (widget && widget.dataset.sitekey) return widget.dataset.sitekey;
    var form = document.querySelector(
      '#resume-modal-inner form[data-turnstile-sitekey]'
    );
    return form ? form.getAttribute("data-turnstile-sitekey") : "";
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

  function loadTurnstileScript() {
    return new Promise(function (resolve, reject) {
      if (window.turnstile) {
        resolve();
        return;
      }
      if (scriptLoading) {
        document.addEventListener(
          "turnstile-loaded",
          function () {
            resolve();
          },
          { once: true }
        );
        return;
      }
      scriptLoading = true;
      window.__resumeTurnstileOnload = function () {
        document.dispatchEvent(new Event("turnstile-loaded"));
        resolve();
      };
      var s = document.createElement("script");
      s.src =
        "https://challenges.cloudflare.com/turnstile/v0/api.js?onload=__resumeTurnstileOnload&render=explicit";
      s.async = true;
      s.defer = true;
      s.setAttribute("data-resume-turnstile", "true");
      s.onerror = function () {
        reject(new Error("Failed to load Turnstile script"));
      };
      document.head.appendChild(s);
    });
  }

  function resetWidget() {
    if (widgetId !== null && window.turnstile) {
      try {
        window.turnstile.remove(widgetId);
      } catch (_) {}
    }
    widgetId = null;
    setToken("");
    var container = document.getElementById("turnstile-widget");
    if (container) container.innerHTML = "";
  }

  function renderWidget() {
    var container = document.getElementById("turnstile-widget");
    if (!container || widgetId !== null) return;

    var siteKey = siteKeyFromDOM();
    if (!siteKey) return;

    loadTurnstileScript()
      .then(function () {
        if (!document.getElementById("turnstile-widget")) return;
        if (widgetId !== null) return;

        widgetId = window.turnstile.render(container, {
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
      })
      .catch(function (err) {
        console.error("Turnstile render failed:", err);
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

  document.body.addEventListener("htmx:beforeRequest", function (evt) {
    var elt = evt.detail && evt.detail.elt;
    if (!elt || elt.getAttribute("hx-post") !== "/api/resume/request") return;

    if (!siteKeyFromDOM()) return;

    var tokenEl = document.getElementById("cf-turnstile-response");
    if (!tokenEl || !tokenEl.value) {
      evt.preventDefault();
      showFormError("Complete the verification challenge before submitting.");
    }
  });
})();
