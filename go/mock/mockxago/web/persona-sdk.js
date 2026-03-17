(function () {
  if (window.Persona && window.Persona.Client) {
    return;
  }

  function MockPersonaClient(options) {
    this.options = options || {};
    this.overlay = null;
    this.onMessage = null;

    var self = this;
    setTimeout(function () {
      if (typeof self.options.onReady === 'function') {
        self.options.onReady();
      }
    }, 0);
  }

  MockPersonaClient.prototype.open = function () {
    var inquiryId = this.options.inquiryId;
    var sessionToken = this.options.sessionToken || '';

    if (!inquiryId) {
      if (typeof this.options.onError === 'function') {
        this.options.onError(new Error('Missing inquiryId'));
      }
      return;
    }

    var script = document.currentScript || (function () {
      var scripts = document.getElementsByTagName('script');
      return scripts[scripts.length - 1];
    })();

    var sdkOrigin = window.location.origin;
    if (script && script.src) {
      try {
        sdkOrigin = new URL(script.src).origin;
      } catch (_) {
      }
    }

    var iframeUrl = sdkOrigin + '/v1/inquiries/' + encodeURIComponent(inquiryId) + '/iframe?token=' + encodeURIComponent(sessionToken);

    var overlay = document.createElement('div');
    overlay.setAttribute('data-testid', 'persona-overlay');
    overlay.style.position = 'fixed';
    overlay.style.top = '0';
    overlay.style.left = '0';
    overlay.style.right = '0';
    overlay.style.bottom = '0';
    overlay.style.background = 'rgba(0, 0, 0, 0.4)';
    overlay.style.zIndex = '2147483647';
    overlay.style.display = 'flex';
    overlay.style.alignItems = 'center';
    overlay.style.justifyContent = 'center';

    var iframe = document.createElement('iframe');
    iframe.setAttribute('title', 'Activate wallet');
    iframe.setAttribute('allow', 'camera;microphone');
    iframe.setAttribute('sandbox', 'allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts');
    iframe.style.width = '100%';
    iframe.style.maxWidth = '700px';
    iframe.style.height = '90vh';
    iframe.style.border = '0';
    iframe.style.borderRadius = '8px';
    iframe.src = iframeUrl;

    overlay.appendChild(iframe);
    document.body.appendChild(overlay);

    this.overlay = overlay;

    var self = this;
    var expectedOrigin = sdkOrigin;
    this.onMessage = function (event) {
      if (event.origin !== expectedOrigin) {
        return;
      }
      if (!event || !event.data || event.data.type !== 'OnboardingCompleted') {
        return;
      }

      var parsedValue;
      try {
        parsedValue = JSON.parse(event.data.value);
      } catch (_) {
        return;
      }

      if (parsedValue && parsedValue.applicantStatus === 'submitted') {
        self.close();
        if (typeof self.options.onComplete === 'function') {
          self.options.onComplete({
            inquiryId: inquiryId,
            status: 'completed',
            fields: {}
          });
        }
      }
    };

    window.addEventListener('message', this.onMessage);
  };

  MockPersonaClient.prototype.close = function () {
    if (this.onMessage) {
      window.removeEventListener('message', this.onMessage);
      this.onMessage = null;
    }

    if (this.overlay && this.overlay.parentNode) {
      this.overlay.parentNode.removeChild(this.overlay);
    }

    this.overlay = null;
  };

  window.Persona = {
    Client: MockPersonaClient
  };
})();
