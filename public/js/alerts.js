/**
 * alerts.js — Fade Alert Helpers
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
 *   Usage: flashMessages()                            // all .flash-message, 4s
 *          flashMessages('.flash-message', { delay: 5000 })
 */

(function (global) {
    const DEFAULT_DELAY = 4000;
    const FADE_IN_MS = 300;
    const FADE_OUT_MS = 500;

    function fadeAlert(target, opts) {
        const el = typeof target === 'string' ? document.querySelector(target) : target;
        if (!el) return;

        const delay = (opts && opts.delay !== undefined) ? opts.delay : DEFAULT_DELAY;
        const auto = (opts && opts.auto !== undefined) ? opts.auto : true;

        // Fade-in
        el.style.opacity = '0';
        el.style.transform = 'translateY(-8px)';
        el.style.transition = `opacity ${FADE_IN_MS}ms ease, transform ${FADE_IN_MS}ms ease`;

        requestAnimationFrame(() => {
            requestAnimationFrame(() => {
                el.style.opacity = '1';
                el.style.transform = 'translateY(0)';
            });
        });

        function dismiss() {
            el.style.transition = `opacity ${FADE_OUT_MS}ms ease`;
            el.style.opacity = '0';
            el.style.pointerEvents = 'none';
            setTimeout(() => el.remove(), FADE_OUT_MS);
        }

        if (auto && delay > 0) {
            setTimeout(dismiss, delay);
        }

        return dismiss;
    }

    /**
     * Apply fadeAlert to every element matching selector (default: '.flash-message').
     * @param {string} [selector='.flash-message']
     * @param {object} [opts] - same as fadeAlert opts
     */
    function flashMessages(selector, opts) {
        const sel = selector || '.flash-message';
        document.querySelectorAll(sel).forEach(function (el) {
            fadeAlert(el, opts);
        });
    }

    global.fadeAlert = fadeAlert;
    global.flashMessages = flashMessages;
})(window);
