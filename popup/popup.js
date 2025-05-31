// SlopSquid Popup Script 🦑

document.addEventListener('DOMContentLoaded', function() {
    const toggleBtn = document.getElementById('toggleBtn');
    const optionsBtn = document.getElementById('optionsBtn');
    const detectedCount = document.getElementById('detectedCount');
    const status = document.getElementById('status');

    // Get current tab
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
        const tab = tabs[0];
        
        // Get current stats from content script
        chrome.tabs.sendMessage(tab.id, {action: 'getStats'}, function(response) {
            if (response) {
                detectedCount.textContent = response.markedElements || 0;
            }
        });
    });

    // Load current settings
    chrome.storage.sync.get({
        enabled: true
    }, function(items) {
        updateUI(items.enabled);
    });

    // Toggle button click
    toggleBtn.addEventListener('click', function() {
        chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
            const tab = tabs[0];
            
            chrome.tabs.sendMessage(tab.id, {action: 'toggle'}, function(response) {
                if (response) {
                    updateUI(response.enabled);
                    
                    // Save the new state
                    chrome.storage.sync.set({
                        enabled: response.enabled
                    });
                }
            });
        });
    });

    // Options button click
    optionsBtn.addEventListener('click', function() {
        chrome.runtime.openOptionsPage();
    });

    function updateUI(enabled) {
        if (enabled) {
            toggleBtn.textContent = 'Disable';
            toggleBtn.className = 'toggle-btn enabled';
            status.textContent = 'Active';
        } else {
            toggleBtn.textContent = 'Enable';
            toggleBtn.className = 'toggle-btn disabled';
            status.textContent = 'Disabled';
        }
    }
}); 