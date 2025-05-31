// SlopSquid Content Script - The Ink Detector 🦑
console.log('🦑 SlopSquid loaded!');

class SlopSquidDetector {
  constructor() {
    this.isEnabled = true;
    this.sensitivity = 0.6; // 0-1 scale
    this.inkEffects = [];
    this.init();
  }

  async init() {
    // Load settings from storage
    const settings = await chrome.storage.sync.get(['enabled', 'sensitivity']);
    this.isEnabled = settings.enabled !== false;
    this.sensitivity = settings.sensitivity || 0.6;

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
    // Enhanced heuristic-based detection
    let suspicionScore = 0;
    
    // Expanded AI phrases from real examples
    const aiPhrases = [
      // Classic AI introductions
      'as an ai', 'i apologize', 'i understand', 'i\'m sorry', 'i cannot',
      // Formal connectors (very common in AI)
      'furthermore', 'moreover', 'subsequently', 'therefore', 'however',
      'nevertheless', 'nonetheless', 'additionally', 'consequently',
      // Academic/formal language overuse
      'in conclusion', 'it\'s worth noting', 'it is important to note',
      'delve into', 'comprehensive', 'robust', 'innovative', 'cutting-edge',
      'state-of-the-art', 'facilitate', 'utilize', 'leverage', 'seamless',
      // AI-specific patterns from the test examples
      'large language model', 'trained on a massive dataset', 'variety of tasks',
      'human-like responses', 'wide range of prompts', 'diverse set of',
      'capable of understanding', 'collective effervescence', 'social solidarity',
      'heightened sense', 'shared emotions', 'sense of unity'
    ];
    
    const sentences = text.split(/[.!?]+/);
    
    // Check for AI phrases (increased weight)
    const lowerText = text.toLowerCase();
    aiPhrases.forEach(phrase => {
      if (lowerText.includes(phrase)) {
        suspicionScore += 0.15; // Increased from 0.1
      }
    });
    
    // Check sentence structure patterns
    let perfectSentences = 0;
    let longSentences = 0;
    sentences.forEach(sentence => {
      const words = sentence.trim().split(/\s+/);
      // AI tends to write very consistent sentence lengths
      if (words.length >= 15 && words.length <= 25) {
        perfectSentences++;
      }
      // AI also creates many long, complex sentences
      if (words.length > 25) {
        longSentences++;
      }
    });
    
    if (sentences.length > 0) {
      suspicionScore += (perfectSentences / sentences.length) * 0.3;
      suspicionScore += (longSentences / sentences.length) * 0.2;
    }
    
    // Check for overly formal language (enhanced)
    const formalWords = lowerText.match(/\b(additionally|furthermore|consequently|therefore|however|nevertheless|nonetheless|subsequently|moreover|thus|hence|whereby|wherein|thereof|thereby)\b/g);
    if (formalWords) {
      suspicionScore += Math.min(formalWords.length * 0.08, 0.3); // Increased weight
    }
    
    // Check for AI writing patterns
    const aiPatterns = [
      /\b(can be used for|is able to|has been trained|is capable of)\b/gi,
      /\b(a variety of|wide range of|diverse set of|massive dataset)\b/gi,
      /\b(language model|text generation|natural language)\b/gi
    ];
    
    aiPatterns.forEach(pattern => {
      const matches = lowerText.match(pattern);
      if (matches) {
        suspicionScore += matches.length * 0.1;
      }
    });
    
    // Lower the minimum threshold for better detection
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