// MockPlaid drop-in replacement for Plaid's hosted SDK.
// Served at https://cdn.plaid.com/link/v2/stable/link-initialize.js when
// cdn.plaid.com is redirected to mockplaid. Defines window.Plaid with the same
// create()/open()/exit()/destroy() contract react-plaid-link expects, and bridges
// the mock Link iframe's postMessage events into onSuccess/onExit.
(function () {
  'use strict';

  var CDN_ORIGIN = 'https://cdn.plaid.com';

  function buildOverlay(src) {
    var overlay = document.createElement('div');
    overlay.id = 'mockplaid-overlay';
    overlay.style.cssText =
      'position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:2147483647;' +
      'display:flex;align-items:center;justify-content:center;';
    var iframe = document.createElement('iframe');
    iframe.id = 'mockplaid-iframe';
    iframe.src = src;
    iframe.style.cssText =
      'width:380px;height:600px;max-width:95vw;max-height:90vh;border:none;' +
      'border-radius:12px;background:#fff;box-shadow:0 10px 40px rgba(0,0,0,0.35);';
    overlay.appendChild(iframe);
    return overlay;
  }

  window.Plaid = {
    create: function (config) {
      config = config || {};
      var overlay = null;
      var messageHandler = null;

      function teardown() {
        if (messageHandler) {
          window.removeEventListener('message', messageHandler);
          messageHandler = null;
        }
        if (overlay && overlay.parentNode) {
          overlay.parentNode.removeChild(overlay);
        }
        overlay = null;
      }

      // react-plaid-link flips `ready` off onLoad — fire it async so the create()
      // call has returned first.
      setTimeout(function () {
        if (typeof config.onLoad === 'function') config.onLoad();
      }, 0);

      return {
        open: function () {
          if (overlay) return;
          overlay = buildOverlay(
            CDN_ORIGIN + '/link?link_token=' + encodeURIComponent(config.token || '')
          );
          messageHandler = function (event) {
            if (event.origin !== CDN_ORIGIN) return;
            var data = event.data || {};
            if (data.namespace !== 'mockplaid') return;
            if (data.type === 'success') {
              teardown();
              if (typeof config.onSuccess === 'function') {
                config.onSuccess(data.public_token, data.metadata);
              }
            } else if (data.type === 'exit') {
              teardown();
              if (typeof config.onExit === 'function') {
                config.onExit(data.error || null, data.metadata || {});
              }
            }
          };
          window.addEventListener('message', messageHandler);
          document.body.appendChild(overlay);
        },
        exit: function (_opts, cb) {
          teardown();
          if (typeof cb === 'function') cb();
        },
        submit: function () {},
        destroy: function () {
          teardown();
        }
      };
    }
  };
})();
