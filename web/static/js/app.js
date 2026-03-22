/**
 * app.js — Go-PostfixAdmin Shared Functions (jQuery 4.0.0)
 *
 * All reusable UI functions extracted from inline scripts.
 * Page-specific translated strings are passed via configuration objects.
 */
var App = (function ($) {
    'use strict';

    function getPasswordI18n() {
        var i18n = window.AppI18n && window.AppI18n.Password;

        return $.extend({
            genFail: 'Failed to generate password. Please try again.',
            weak: 'Weak',
            medium: 'Medium',
            good: 'Good',
            strong: 'Strong',
            match: '✓ Passwords match',
            noMatch: '✗ Passwords do not match',
            minChars: 'Minimum 9 characters',
            uppercase: 'Must include at least one uppercase letter',
            lowercase: 'Must include at least one lowercase letter',
            number: 'Must include at least one number',
            special: 'Must include at least one special character (!@#$%...)',
            minLen: 'The password must be at least 8 characters long',
            noMatchAlert: 'Passwords do not match'
        }, i18n || {});
    }

    function resolvePasswordLabels(labels) {
        var i18n = getPasswordI18n();

        return $.extend({
            weak: i18n.weak,
            medium: i18n.medium,
            good: i18n.good,
            strong: i18n.strong,
            match: i18n.match,
            noMatch: i18n.noMatch,
            passwordMinChars: i18n.minChars,
            passwordUppercase: i18n.uppercase,
            passwordLowercase: i18n.lowercase,
            passwordNumber: i18n.number,
            passwordSpecial: i18n.special
        }, labels || {});
    }

    function resolvePasswordValidationMsgs(msgs) {
        var i18n = getPasswordI18n();

        return $.extend({
            minLen: i18n.minLen,
            noMatch: i18n.noMatchAlert
        }, msgs || {});
    }

    function buildPasswordOpts(opts) {
        var base = opts || {};

        return $.extend({}, base, {
            labels: resolvePasswordLabels(base.labels),
            validationMsgs: resolvePasswordValidationMsgs(base.validationMsgs),
            failMsg: base.failMsg || getPasswordI18n().genFail
        });
    }

    // ─── Toggle Password Visibility ──────────────────────────────
    function togglePassword(fieldId, btn) {
        var $field = $('#' + fieldId);
        var $icon = $(btn).find('i');

        if ($field.attr('type') === 'password') {
            $field.attr('type', 'text');
            $icon.attr('data-lucide', 'eye-off');
        } else {
            $field.attr('type', 'password');
            $icon.attr('data-lucide', 'eye');
        }
        lucide.createIcons();
    }

    // ─── Toggle Domains Card (Superadmin checkbox) ───────────────
    function toggleDomains() {
        var isSuper = $('#superadmin').is(':checked');
        var $card = $('#domainsCard');

        if (isSuper) {
            $card.addClass('opacity-50 pointer-events-none');
        } else {
            $card.removeClass('opacity-50 pointer-events-none');
        }
    }

    // ─── Password Strength Checker ───────────────────────────────
    // opts: { passwordId, meterId, barId, textId, labels: { weak, medium, good, strong } }
    function checkPasswordStrength(opts) {
        var labels = resolvePasswordLabels(opts.labels);
        var password = $('#' + opts.passwordId).val();
        var $meter = $('#' + opts.meterId);
        var $bar = $('#' + opts.barId);
        var $text = $('#' + opts.textId);

        if (password.length === 0) {
            $meter.addClass('hidden');
            return;
        }

        $meter.removeClass('hidden');

        var strength = 0;
        if (password.length >= 8) strength += 25;
        if (password.length >= 12) strength += 15;
        if (/[a-z]/.test(password)) strength += 15;
        if (/[A-Z]/.test(password)) strength += 15;
        if (/[0-9]/.test(password)) strength += 15;
        if (/[^a-zA-Z0-9]/.test(password)) strength += 15;

        var label, color;
        if (strength < 40) {
            label = labels.weak;
            color = '#DC2626';
        } else if (strength < 60) {
            label = labels.medium;
            color = '#F59E0B';
        } else if (strength < 80) {
            label = labels.good;
            color = '#10B981';
        } else {
            label = labels.strong;
            color = '#059669';
        }

        $bar.css({ width: strength + '%', backgroundColor: color });
        $text.text(label).css('color', color);
    }

    /**
     * Retorna uma mensagem de erro se a senha nao atender aos requisitos,
     * ou `null` se estiver ok.
     */
    function passwordMsg(password, labels) {
        var msgs = resolvePasswordLabels(labels);

        if (password.length < 9) {
            return msgs.passwordMinChars || 'Minimum 9 characters';
        }
        if (!/[A-Z]/.test(password)) {
            return msgs.passwordUppercase || 'Must include at least one uppercase letter';
        }
        if (!/[a-z]/.test(password)) {
            return msgs.passwordLowercase || 'Must include at least one lowercase letter';
        }
        if (!/[0-9]/.test(password)) {
            return msgs.passwordNumber || 'Must include at least one number';
        }
        if (!/[^A-Za-z0-9]/.test(password)) {
            return msgs.passwordSpecial || 'Must include at least one special character (!@#$%...)';
        }

        return null;
    }

    // ─── Password Match Checker ──────────────────────────────────
    // opts: { passwordId, confirmId, indicatorId, submitBtnId?, labels: { match, noMatch, ...password validation labels } }
    function checkPasswordMatch(opts) {
        var labels = resolvePasswordLabels(opts.labels);
        var password = $('#' + opts.passwordId).val();
        var confirm = $('#' + opts.confirmId).val();
        var $indicator = $('#' + opts.indicatorId);
        var $submit = opts.submitBtnId ? $('#' + opts.submitBtnId) : null;
        var err = passwordMsg(password, labels);

        if (confirm.length === 0) {
            $indicator.addClass('hidden');
            return;
        }

        $indicator.removeClass('hidden');

        if (err) {
            $indicator.text(err)
                .attr('class', 'text-xs mt-2 font-bold text-red-600');
            if ($submit) {
                $submit.prop('disabled', true)
                    .addClass('opacity-50 cursor-not-allowed');
            }
            return;
        }

        if (password === confirm) {
            $indicator.text(labels.match)
                .attr('class', 'text-xs mt-2 font-bold text-green-600');
            if ($submit) {
                $submit.prop('disabled', false)
                    .removeClass('opacity-50 cursor-not-allowed');
            }
        } else {
            $indicator.text(labels.noMatch)
                .attr('class', 'text-xs mt-2 font-bold text-red-600');
            if ($submit) {
                $submit.prop('disabled', true)
                    .addClass('opacity-50 cursor-not-allowed');
            }
        }
    }

    // ─── Generate Password ───────────────────────────────────────
    // opts: { passwordId, confirmId, onSuccess (callback), failMsg }
    function generatePassword(opts) {
        var upperChars = 'ABCDEFGHJKLMNPQRSTUVWXYZ';
        var lowerChars = 'abcdefghijkmnpqrstuvwxyz';
        var digitChars = '23456789';
        var specialChars = '!@#$%&*';
        var allChars = upperChars + lowerChars + digitChars + specialChars;
        var length = 14;

        function randomChar(chars) {
            return chars[Math.floor(Math.random() * chars.length)];
        }

        function shuffle(chars) {
            for (var i = chars.length - 1; i > 0; i--) {
                var j = Math.floor(Math.random() * (i + 1));
                var tmp = chars[i];
                chars[i] = chars[j];
                chars[j] = tmp;
            }
            return chars;
        }

        var pwdChars = [
            randomChar(upperChars),
            randomChar(lowerChars),
            randomChar(digitChars),
            randomChar(specialChars)
        ];

        while (pwdChars.length < length) {
            pwdChars.push(randomChar(allChars));
        }

        var pwd = shuffle(pwdChars).join('');

        var $password = $('#' + opts.passwordId);
        var $confirm = $('#' + opts.confirmId);

        $password.attr('type', 'text');
        $confirm.attr('type', 'text');
        $password.siblings('button').find('i').attr('data-lucide', 'eye-off');
        $confirm.siblings('button').find('i').attr('data-lucide', 'eye-off');

        $password.val(pwd).attr('value', pwd).trigger('input').trigger('change');
        $confirm.val(pwd).attr('value', pwd).trigger('input').trigger('change');
        lucide.createIcons();

        if (typeof opts.onSuccess === 'function') {
            opts.onSuccess(pwd);
        }
    }

    // ─── Confirm Delete Resource ─────────────────────────────────
    // opts: { url, replacements?, msgs: { confirm, success, error, requestError } }
    function confirmDeleteResource(opts) {
        var confirmMsg = opts.msgs.confirm;
        if (opts.replacements) {
            $.each(opts.replacements, function (key, val) {
                confirmMsg = confirmMsg.split('${' + key + '}').join(val);
            });
        }
        if (!confirm(confirmMsg)) {
            return;
        }

        $.ajax({
            url: opts.url,
            method: 'DELETE',
            contentType: 'application/json'
        }).done(function (data) {
            if (data.success) {
                // Success message will be set in flash session by backend and shown on reload.
                window.location.reload();
            } else {
                showToast('Error', opts.msgs.error + (data.error || 'Unknown error'), 'error');
            }
        }).fail(function (err) {
            showToast('Error', opts.msgs.requestError + err.statusText, 'error');
        });
    }

    /**
     * Show a toast notification.
     * @param {string} heading - Toast heading (Success / Error).
     * @param {string} text - Toast body text.
     * @param {string} icon - Icon type: 'success' | 'error' | 'warning' | 'info'.
     */
    function showToast(heading, text, icon) {
        $.toast({
            heading: heading,
            text: text,
            showHideTransition: 'slide',
            icon: icon || 'info', // O icone determina a cor gerada pela biblioteca.
            position: 'top-right',
            hideAfter: 3000,
            allowToastClose: true,
        });
    }

    // ─── Auto-dismiss Flash Messages (FadeAlert) ─────────────────
    var ALERT_DEFAULT_DELAY = 4000;
    var ALERT_FADE_IN_MS = 300;
    var ALERT_FADE_OUT_MS = 500;

    function fadeAlert(target, opts) {
        var $el = (typeof target === 'string') ? $(target) : $(target);
        if (!$el.length) return;

        var delay = (opts && opts.delay !== undefined) ? opts.delay : ALERT_DEFAULT_DELAY;
        var auto = (opts && opts.auto !== undefined) ? opts.auto : true;

        // Fade-in
        $el.css({
            opacity: '0',
            transform: 'translateY(-8px)',
            transition: 'opacity ' + ALERT_FADE_IN_MS + 'ms ease, transform ' + ALERT_FADE_IN_MS + 'ms ease'
        });

        requestAnimationFrame(function () {
            requestAnimationFrame(function () {
                $el.css({ opacity: '1', transform: 'translateY(0)' });
            });
        });

        function dismiss() {
            $el.css({
                transition: 'opacity ' + ALERT_FADE_OUT_MS + 'ms ease',
                opacity: '0',
                pointerEvents: 'none'
            });
            setTimeout(function () { $el.remove(); }, ALERT_FADE_OUT_MS);
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

    // ─── Check Password Change Intention ─────────────────────────
    // opts: { passwordId, confirmId, changeInputId, meterId, indicatorId }
    function checkPasswordChangeIntention(opts) {
        var password = $('#' + opts.passwordId).val();
        var confirm = $('#' + opts.confirmId).val();
        var $change = $('#' + opts.changeInputId);

        if (password.length > 0 || confirm.length > 0) {
            $change.val('true');
            $('#' + opts.passwordId).prop('required', true);
            $('#' + opts.confirmId).prop('required', true);
        } else {
            $change.val('false');
            $('#' + opts.passwordId).prop('required', false);
            $('#' + opts.confirmId).prop('required', false);
            $('#' + opts.meterId).addClass('hidden');
            $('#' + opts.indicatorId).addClass('hidden');
        }
    }

    // ─── Email Preview (Add Mailbox) ─────────────────────────────
    function updateEmailPreview() {
        var localPart = ($('#local_part').val() || '').toLowerCase() || 'user';
        var domain = $('#domain').val() || 'domain.com';
        $('#localPartPreview').text(localPart);
        $('#domainPreview').text(domain);
    }

    // ─── Email Validation ────────────────────────────────────────
    function validateEmail(email) {
        return String(email)
            .toLowerCase()
            .match(
                /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
            );
    }

    // ─── Init Password Form ──────────────────────────────────────
    // Wires up event listeners for password fields with strength + match checking
    // opts: { passwordId, confirmId, meterId, barId, textId, indicatorId, submitBtnId?,
    //         formId?, changeInputId?, labels: { weak, medium, good, strong, match, noMatch },
    //         validationMsgs?: { minLen, noMatch }, minLen? }
    function initPasswordForm(opts) {
        var labels = resolvePasswordLabels(opts.labels);
        var validationMsgs = resolvePasswordValidationMsgs(opts.validationMsgs);
        var strengthOpts = {
            passwordId: opts.passwordId,
            meterId: opts.meterId,
            barId: opts.barId,
            textId: opts.textId,
            labels: labels
        };

        var matchOpts = {
            passwordId: opts.passwordId,
            confirmId: opts.confirmId,
            indicatorId: opts.indicatorId,
            submitBtnId: opts.submitBtnId,
            labels: {
                match: labels.match,
                noMatch: labels.noMatch,
                passwordMinChars: labels.passwordMinChars,
                passwordUppercase: labels.passwordUppercase,
                passwordLowercase: labels.passwordLowercase,
                passwordNumber: labels.passwordNumber,
                passwordSpecial: labels.passwordSpecial
            }
        };

        var intentionOpts = opts.changeInputId ? {
            passwordId: opts.passwordId,
            confirmId: opts.confirmId,
            changeInputId: opts.changeInputId,
            meterId: opts.meterId,
            indicatorId: opts.indicatorId
        } : null;

        $('#' + opts.passwordId).on('input', function () {
            if (intentionOpts) checkPasswordChangeIntention(intentionOpts);
            checkPasswordStrength(strengthOpts);
            checkPasswordMatch(matchOpts);
        });

        $('#' + opts.confirmId).on('input', function () {
            if (intentionOpts) checkPasswordChangeIntention(intentionOpts);
            checkPasswordMatch(matchOpts);
        });

        // Form validation on submit
        if (opts.formId) {
            $('#' + opts.formId).on('submit', function (e) {
                var isEdit = opts.changeInputId && $('#' + opts.changeInputId).val() !== 'true';
                if (isEdit) return; // Skip validation if not changing password

                var password = $('#' + opts.passwordId).val();
                var confirm = $('#' + opts.confirmId).val();
                var minLen = opts.minLen || 8;

                if (password.length < minLen) {
                    e.preventDefault();
                    alert(validationMsgs.minLen);
                    return false;
                }

                if (password !== confirm) {
                    e.preventDefault();
                    alert(validationMsgs.noMatch);
                    return false;
                }
            });
        }
    }

    // ─── Public API ──────────────────────────────────────────────
    return {
        togglePassword: togglePassword,
        toggleDomains: toggleDomains,
        checkPasswordStrength: checkPasswordStrength,
        checkPasswordMatch: checkPasswordMatch,
        passwordMsg: passwordMsg,
        generatePassword: generatePassword,
        confirmDeleteResource: confirmDeleteResource,
        fadeAlert: fadeAlert,
        flashMessages: flashMessages,
        showToast: showToast,
        checkPasswordChangeIntention: checkPasswordChangeIntention,
        updateEmailPreview: updateEmailPreview,
        validateEmail: validateEmail,
        initPasswordForm: initPasswordForm,
        buildPasswordOpts: buildPasswordOpts
    };

})(jQuery);
