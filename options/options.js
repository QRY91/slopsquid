// SlopSquid Options Script 🦑

document.addEventListener('DOMContentLoaded', function() {
    const enabledCheckbox = document.getElementById('enabled');
    const sensitivitySlider = document.getElementById('sensitivity');
    const sensitivityValue = document.getElementById('sensitivityValue');
    const notificationsCheckbox = document.getElementById('notifications');
    const visualEffectsCheckbox = document.getElementById('visualEffects');
    const status = document.getElementById('status');

    // Load saved settings
    chrome.storage.sync.get({
        enabled: true,
        sensitivity: 70,
        showNotifications: true,
        visualEffects: true
    }, function(items) {
        enabledCheckbox.checked = items.enabled;
        sensitivitySlider.value = items.sensitivity;
        sensitivityValue.textContent = items.sensitivity + '%';
        notificationsCheckbox.checked = items.showNotifications;
        visualEffectsCheckbox.checked = items.visualEffects;
    });

    // Update sensitivity value display
    sensitivitySlider.addEventListener('input', function() {
        sensitivityValue.textContent = this.value + '%';
        saveSettings();
    });

    // Save settings when changed
    enabledCheckbox.addEventListener('change', saveSettings);
    notificationsCheckbox.addEventListener('change', saveSettings);
    visualEffectsCheckbox.addEventListener('change', saveSettings);

    function saveSettings() {
        const settings = {
            enabled: enabledCheckbox.checked,
            sensitivity: parseInt(sensitivitySlider.value),
            showNotifications: notificationsCheckbox.checked,
            visualEffects: visualEffectsCheckbox.checked
        };

        chrome.storage.sync.set(settings, function() {
            // Show save confirmation
            status.textContent = '✅ Settings saved!';
            status.style.background = 'rgba(0, 255, 0, 0.3)';
            
            setTimeout(function() {
                status.textContent = 'Settings saved automatically';
                status.style.background = 'rgba(0, 255, 0, 0.2)';
            }, 2000);
        });

        // Notify all tabs of settings change
        chrome.tabs.query({}, function(tabs) {
            tabs.forEach(function(tab) {
                chrome.tabs.sendMessage(tab.id, {
                    action: 'settingsChanged',
                    settings: settings
                }).catch(() => {
                    // Ignore errors for tabs without content script
                });
            });
        });
    }
}); 