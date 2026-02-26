/**
 * alerts.js — Fade Alert Helpers (jQuery 4.0.0)
 *
 * fadeAlert(target, opts?)
 *   Applies fade-in + auto-dismiss to a single element.
 *   @param {string|HTMLElement} target  - CSS selector or DOM element
 *   @param {object} [opts]
 *   @param {number}  [opts.delay=4000]  - ms before auto-dismiss (0 = no auto)
 *   @param {boolean} [opts.auto=true]   - enable auto-dismiss
 *   @returns {Function} dismiss — call it to manually fade-out
 *
 * flashMessages(selector?, opts?)
 *   Applies fadeAlert to ALL elements matching the selector (default: '.flash-message').
 */

(function (global, $) {
    var DEFAULT_DELAY = 4000;
    var FADE_IN_MS = 300;
    var FADE_OUT_MS = 500;

    function fadeAlert(target, opts) {
        var $el = (typeof target === 'string') ? $(target) : $(target);
        if (!$el.length) return;

        var delay = (opts && opts.delay !== undefined) ? opts.delay : DEFAULT_DELAY;
        var auto = (opts && opts.auto !== undefined) ? opts.auto : true;

        // Fade-in
        $el.css({
            opacity: '0',
            transform: 'translateY(-8px)',
            transition: 'opacity ' + FADE_IN_MS + 'ms ease, transform ' + FADE_IN_MS + 'ms ease'
        });

        requestAnimationFrame(function () {
            requestAnimationFrame(function () {
                $el.css({ opacity: '1', transform: 'translateY(0)' });
            });
        });

        function dismiss() {
            $el.css({
                transition: 'opacity ' + FADE_OUT_MS + 'ms ease',
                opacity: '0',
                pointerEvents: 'none'
            });
            setTimeout(function () { $el.remove(); }, FADE_OUT_MS);
        }

        if (auto && delay > 0) {
            setTimeout(dismiss, delay);
        }

        return dismiss;
    }

    function flashMessages(selector, opts) {
        var sel = selector || '.flash-message';
        $(sel).each(function () {
            fadeAlert(this, opts);
        });
    }

    global.fadeAlert = fadeAlert;
    global.flashMessages = flashMessages;
})(window, jQuery);
