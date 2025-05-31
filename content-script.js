// SlopSquid Content Script - The Ink Detector 🦑
console.log('🦑 SlopSquid loaded!');

class SlopSquidDetector {
  constructor() {
    this.isEnabled = true;
    this.sensitivity = 0.7; // 0-1 scale
    this.inkEffects = [];
    this.init();
  }

  async init() {
    // Load settings from storage
    const settings = await chrome.storage.sync.get(['enabled', 'sensitivity']);
    this.isEnabled = settings.enabled !== false;
    this.sensitivity = settings.sensitivity || 0.7;

    if (this.isEnabled) {
      this.scanPage();
      this.observeChanges();
    }
  }

  scanPage() {
    // Find all text nodes that could contain AI-generated content
    const textElements = document.querySelectorAll('p, div, span, article, section, h1, h2, h3, h4, h5, h6');
    
    textElements.forEach(element => {
      const text = element.textContent.trim();
      if (text.length > 50) { // Only analyze substantial text
        this.analyzeText(element, text);
      }
    });
  }

  analyzeText(element, text) {
    const confidence = this.detectAIContent(text);
    
    if (confidence > this.sensitivity) {
      this.applyInkEffect(element, confidence);
    }
  }

  detectAIContent(text) {
    // Simple heuristic-based detection (we'll enhance this)
    let suspicionScore = 0;
    
    // Common AI patterns
    const aiPhrases = [
      'as an ai', 'i apologize', 'i understand', 'furthermore', 'moreover',
      'in conclusion', 'it\'s worth noting', 'delve into', 'comprehensive',
      'leverage', 'utilize', 'facilitate', 'subsequently', 'robust',
      'seamless', 'innovative', 'cutting-edge', 'state-of-the-art'
    ];
    
    const sentences = text.split(/[.!?]+/);
    
    // Check for AI phrases
    aiPhrases.forEach(phrase => {
      if (text.toLowerCase().includes(phrase)) {
        suspicionScore += 0.1;
      }
    });
    
    // Check sentence structure patterns
    let perfectSentences = 0;
    sentences.forEach(sentence => {
      const words = sentence.trim().split(/\s+/);
      // AI tends to write very consistent sentence lengths
      if (words.length >= 15 && words.length <= 25) {
        perfectSentences++;
      }
    });
    
    if (sentences.length > 0) {
      suspicionScore += (perfectSentences / sentences.length) * 0.3;
    }
    
    // Check for overly formal language
    const formalWords = text.toLowerCase().match(/\b(additionally|furthermore|consequently|therefore|however|nevertheless|nonetheless|subsequently)\b/g);
    if (formalWords) {
      suspicionScore += Math.min(formalWords.length * 0.05, 0.2);
    }
    
    return Math.min(suspicionScore, 1.0);
  }

  applyInkEffect(element, confidence) {
    // Don't double-mark elements
    if (element.classList.contains('slopsquid-marked')) return;
    
    element.classList.add('slopsquid-marked');
    element.dataset.slopConfidence = confidence.toFixed(2);
    
    // Apply CSS class based on confidence level
    if (confidence > 0.9) {
      element.classList.add('slopsquid-high');
    } else if (confidence > 0.7) {
      element.classList.add('slopsquid-medium');
    } else {
      element.classList.add('slopsquid-low');
    }
    
    // Add tooltip
    element.title = `🦑 SlopSquid detected AI content (${Math.round(confidence * 100)}% confidence)`;
    
    console.log(`🦑 Marked element with ${Math.round(confidence * 100)}% AI confidence`);
  }

  observeChanges() {
    // Watch for dynamically added content
    const observer = new MutationObserver(mutations => {
      mutations.forEach(mutation => {
        mutation.addedNodes.forEach(node => {
          if (node.nodeType === Node.ELEMENT_NODE) {
            const textElements = node.querySelectorAll ? 
              node.querySelectorAll('p, div, span, article, section, h1, h2, h3, h4, h5, h6') : 
              [];
            
            textElements.forEach(element => {
              const text = element.textContent.trim();
              if (text.length > 50 && !element.classList.contains('slopsquid-marked')) {
                this.analyzeText(element, text);
              }
            });
          }
        });
      });
    });
    
    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  toggle() {
    this.isEnabled = !this.isEnabled;
    if (this.isEnabled) {
      this.scanPage();
    } else {
      this.clearEffects();
    }
  }

  clearEffects() {
    document.querySelectorAll('.slopsquid-marked').forEach(element => {
      element.classList.remove('slopsquid-marked', 'slopsquid-high', 'slopsquid-medium', 'slopsquid-low');
      element.removeAttribute('title');
      delete element.dataset.slopConfidence;
    });
  }
}

// Initialize the detector
const slopSquid = new SlopSquidDetector();

// Listen for messages from popup/background
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'toggle') {
    slopSquid.toggle();
    sendResponse({ enabled: slopSquid.isEnabled });
  } else if (request.action === 'getStats') {
    const marked = document.querySelectorAll('.slopsquid-marked').length;
    sendResponse({ markedElements: marked });
  }
}); 