(function () {
  var CLOSE_MS = 280;

  function root() {
    return document.getElementById("resume-modal-root");
  }

  function overlay() {
    return document.getElementById("resume-modal-overlay");
  }

  window.closeResumeModal = function () {
    var el = overlay();
    var modalRoot = root();
    if (!el || !modalRoot || !modalRoot.firstElementChild) return;
    if (el.classList.contains("resume-modal-closing")) return;

    el.classList.add("resume-modal-closing");
    window.setTimeout(function () {
      modalRoot.innerHTML = "";
      document.body.dispatchEvent(new Event("resume-modal-closed"));
    }, CLOSE_MS);
  };

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") window.closeResumeModal();
  });
})();
