/**
 * app.js — Go-PostfixAdmin Shared Functions (jQuery 4.0.0)
 *
 * All reusable UI functions extracted from inline scripts.
 * Page-specific translated strings are passed via configuration objects.
 */
var App = (function ($) {
    'use strict';

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
            label = opts.labels.weak;
            color = '#DC2626';
        } else if (strength < 60) {
            label = opts.labels.medium;
            color = '#F59E0B';
        } else if (strength < 80) {
            label = opts.labels.good;
            color = '#10B981';
        } else {
            label = opts.labels.strong;
            color = '#059669';
        }

        $bar.css({ width: strength + '%', backgroundColor: color });
        $text.text(label).css('color', color);
    }

    // ─── Password Match Checker ──────────────────────────────────
    // opts: { passwordId, confirmId, indicatorId, submitBtnId?, labels: { match, noMatch } }
    function checkPasswordMatch(opts) {
        var password = $('#' + opts.passwordId).val();
        var confirm = $('#' + opts.confirmId).val();
        var $indicator = $('#' + opts.indicatorId);
        var $submit = opts.submitBtnId ? $('#' + opts.submitBtnId) : null;

        if (confirm.length === 0) {
            $indicator.addClass('hidden');
            return;
        }

        $indicator.removeClass('hidden');

        if (password === confirm) {
            $indicator.text(opts.labels.match)
                .attr('class', 'text-xs mt-2 font-bold text-green-600');
            if ($submit) {
                $submit.prop('disabled', false)
                    .removeClass('opacity-50 cursor-not-allowed');
            }
        } else {
            $indicator.text(opts.labels.noMatch)
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
        $.ajax({
            url: '/api/generate-password',
            method: 'GET',
            dataType: 'json'
        }).done(function (data) {
            $('#' + opts.passwordId).val(data.password);
            $('#' + opts.confirmId).val(data.password);

            if (typeof opts.onSuccess === 'function') {
                opts.onSuccess(data.password);
            }
        }).fail(function (err) {
            console.error('Failed to generate password:', err);
            alert(opts.failMsg || 'Failed to generate password');
        });
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
                alert(opts.msgs.success);
                window.location.reload();
            } else {
                alert(opts.msgs.error + (data.error || 'Unknown error'));
            }
        }).fail(function (err) {
            alert(opts.msgs.requestError + err.statusText);
        });
    }

    // ─── Auto-dismiss Flash Messages ─────────────────────────────
    function initFlashMessages(selector, delay) {
        var sel = selector || '.flash-message';
        var ms = delay || 5000;

        $(sel).each(function () {
            var $el = $(this);
            setTimeout(function () {
                $el.animate({ opacity: 0 }, 500, function () {
                    $el.remove();
                });
            }, ms);
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
        var strengthOpts = {
            passwordId: opts.passwordId,
            meterId: opts.meterId,
            barId: opts.barId,
            textId: opts.textId,
            labels: opts.labels
        };

        var matchOpts = {
            passwordId: opts.passwordId,
            confirmId: opts.confirmId,
            indicatorId: opts.indicatorId,
            submitBtnId: opts.submitBtnId,
            labels: {
                match: opts.labels.match,
                noMatch: opts.labels.noMatch
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
        if (opts.formId && opts.validationMsgs) {
            $('#' + opts.formId).on('submit', function (e) {
                var isEdit = opts.changeInputId && $('#' + opts.changeInputId).val() !== 'true';
                if (isEdit) return; // Skip validation if not changing password

                var password = $('#' + opts.passwordId).val();
                var confirm = $('#' + opts.confirmId).val();
                var minLen = opts.minLen || 8;

                if (password.length < minLen) {
                    e.preventDefault();
                    alert(opts.validationMsgs.minLen);
                    return false;
                }

                if (password !== confirm) {
                    e.preventDefault();
                    alert(opts.validationMsgs.noMatch);
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
        generatePassword: generatePassword,
        confirmDeleteResource: confirmDeleteResource,
        initFlashMessages: initFlashMessages,
        checkPasswordChangeIntention: checkPasswordChangeIntention,
        updateEmailPreview: updateEmailPreview,
        validateEmail: validateEmail,
        initPasswordForm: initPasswordForm
    };

})(jQuery);
