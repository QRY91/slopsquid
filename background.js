// SlopSquid Background Service Worker 🦑

chrome.runtime.onInstalled.addListener(() => {
  console.log('🦑 SlopSquid installed!');
  
  // Set default settings
  chrome.storage.sync.set({
    enabled: true,
    sensitivity: 0.7,
    showNotifications: true
  });
});

// Handle extension icon click
chrome.action.onClicked.addListener(async (tab) => {
  // Toggle the extension on current tab
  try {
    const response = await chrome.tabs.sendMessage(tab.id, { action: 'toggle' });
    
    // Update icon based on state
    const iconPath = response.enabled ? 
      'icons/icon32.png' : 
      'icons/icon32-disabled.png';
    
    chrome.action.setIcon({ 
      path: iconPath,
      tabId: tab.id 
    });
    
  } catch (error) {
    console.log('🦑 Could not communicate with content script:', error);
  }
});

// Listen for tab updates to reset icon
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete') {
    // Reset icon to default
    chrome.action.setIcon({ 
      path: 'icons/icon32.png',
      tabId: tabId 
    });
  }
});

// Handle messages from content script and popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'updateBadge') {
    // Show number of detected elements on extension badge
    const text = request.count > 0 ? request.count.toString() : '';
    chrome.action.setBadgeText({
      text: text,
      tabId: sender.tab?.id
    });
    chrome.action.setBadgeBackgroundColor({ color: '#ff0064' });
  }
  
  sendResponse({ received: true });
});

// Context menu for detected text
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: 'slopsquid-analyze',
    title: '🦑 Analyze with SlopSquid',
    contexts: ['selection']
  });
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === 'slopsquid-analyze') {
    // Send selected text to content script for analysis
    chrome.tabs.sendMessage(tab.id, {
      action: 'analyzeSelection',
      text: info.selectionText
    });
  }
}); 