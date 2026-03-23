(function ($) {
    'use strict';

    function showPasswordError(message) {
        $('#passwordErrorText').text(message || 'Request failed');
        $('#passwordError').removeClass('hidden');
        lucide.createIcons();
    }

    function hidePasswordError() {
        $('#passwordError').addClass('hidden');
        $('#passwordErrorText').text('');
    }

    function initUserDashboard() {
        var $passwordForm = $('#passwordForm');
        var $forwardingForm = $('#forwardingForm');

        if (!$passwordForm.length && !$forwardingForm.length) {
            return;
        }

        App.flashMessages();

        if ($passwordForm.length) {
            var passwordOpts = App.buildPasswordOpts({
                passwordId: 'new_password',
                meterId: 'strengthMeter',
                barId: 'strengthBar',
                textId: 'strengthText',
                confirmId: 'confirm_password',
                indicatorId: 'matchIndicator',
                submitBtnId: 'submitBtn'
            });

            App.initPasswordForm($.extend({}, passwordOpts, {
                formId: 'passwordForm'
            }));

            $('#current_password, #new_password, #confirm_password').on('input', hidePasswordError);

            $passwordForm.on('submit', function (e) {
                e.preventDefault();

                var $form = $(this);
                var $submitBtn = $('#submitBtn');

                hidePasswordError();

                if ($submitBtn.prop('disabled')) {
                    return;
                }

                $submitBtn.prop('disabled', true).addClass('opacity-50 cursor-not-allowed');

                $.ajax({
                    url: '/users/password',
                    method: 'POST',
                    data: $form.serialize(),
                    headers: { 'Accept': 'application/json' }
                }).done(function (res) {
                    if (res.success) {
                        App.showToast('Success', res.message || 'Password updated successfully', 'success');
                        $('#current_password, #new_password, #confirm_password').val('');
                        $('#strengthMeter, #matchIndicator').addClass('hidden');
                        $('#strengthBar').css('width', '0%');
                        $('#strengthText').text('');
                        if (res.redirect) {
                            setTimeout(function () {
                                window.location.href = res.redirect;
                            }, 800);
                        }
                    } else {
                        showPasswordError(res.error || 'Request failed');
                        App.showToast('Error', res.error || 'Request failed', 'error');
                    }
                }).fail(function (xhr) {
                    var msg = 'Request failed';
                    try { msg = JSON.parse(xhr.responseText).error || msg; } catch (err) {}
                    showPasswordError(msg);
                    App.showToast('Error', msg, 'error');
                }).always(function () {
                    $submitBtn.prop('disabled', false).removeClass('opacity-50 cursor-not-allowed');
                });
            });
        }

        if ($forwardingForm.length) {
            var $fwdTextarea = $('#forwarding');
            var $fwdBtn = $('#saveForwardingBtn');

            function checkForwarding() {
                var lines = $fwdTextarea.val().trim().split('\n').filter(function (line) {
                    return line.trim() !== '';
                });
                var empty = lines.length === 0;
                var allValid = lines.every(function (line) {
                    return App.validateEmail(line.trim());
                });

                $fwdBtn.prop('disabled', empty || !allValid);
                if (empty || !allValid) {
                    $fwdBtn.addClass('opacity-50 cursor-not-allowed');
                    if (lines.length > 0 && !allValid) {
                        $fwdTextarea.addClass('border-red-500');
                    } else {
                        $fwdTextarea.removeClass('border-red-500');
                    }
                } else {
                    $fwdBtn.removeClass('opacity-50 cursor-not-allowed');
                    $fwdTextarea.removeClass('border-red-500');
                }
            }

            $fwdTextarea.on('input', checkForwarding);
            checkForwarding();

            $forwardingForm.on('submit', function (e) {
                e.preventDefault();

                var $form = $(this);

                if ($fwdBtn.prop('disabled')) {
                    return;
                }

                $fwdBtn.prop('disabled', true).addClass('opacity-50 cursor-not-allowed');

                $.ajax({
                    url: '/users/forwarding',
                    method: 'POST',
                    data: $form.serialize(),
                    headers: { 'Accept': 'application/json' }
                }).done(function (res) {
                    if (res.success) {
                        App.showToast('Success', res.message || 'Forwarding updated successfully', 'success');
                    } else {
                        App.showToast('Error', res.error || 'Request failed', 'error');
                    }
                }).fail(function (xhr) {
                    var msg = 'Request failed';
                    try { msg = JSON.parse(xhr.responseText).error || msg; } catch (err) {}
                    App.showToast('Error', msg, 'error');
                }).always(function () {
                    checkForwarding();
                });
            });
        }
    }

    window.togglePassword = function (fieldId) {
        App.togglePassword(fieldId, event.currentTarget);
    };

    window.generatePassword = function () {
        var passwordOpts = App.buildPasswordOpts({
            passwordId: 'new_password',
            meterId: 'strengthMeter',
            barId: 'strengthBar',
            textId: 'strengthText',
            confirmId: 'confirm_password',
            indicatorId: 'matchIndicator',
            submitBtnId: 'submitBtn'
        });

        App.generatePassword({
            passwordId: 'new_password',
            confirmId: 'confirm_password',
            onSuccess: function () {
                App.checkPasswordStrength(passwordOpts);
                App.checkPasswordMatch(passwordOpts);
            }
        });
    };

    $(function () {
        initUserDashboard();
    });
})(jQuery);
