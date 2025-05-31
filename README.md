# SlopSquid 🦑

**Detect and visualize AI-generated text with inky effects**

*"Slap that AI!"* - A browser extension that marks suspicious AI-generated content with beautiful ink-like visual effects.

## Features

- 🔍 **Real-time AI Detection** - Analyzes text on web pages as you browse
- 🎨 **Ink Effects** - Beautiful visual feedback with confidence-based styling
- ⚡ **Lightweight** - Runs locally with no external API calls
- 🎚️ **Adjustable Sensitivity** - Tune detection to your preferences
- 🔒 **Privacy-First** - No data leaves your browser

## Detection Capabilities

SlopSquid uses heuristic analysis to identify common AI writing patterns:

- **Language Patterns** - Overly formal or repetitive phrasing
- **Sentence Structure** - Consistent length and complexity patterns
- **AI Phrases** - Common AI-generated expressions and transitions
- **Writing Style** - Unnatural formality and corporate speak

## Visual Feedback

- 🟢 **Low Confidence** (70-79%) - Subtle pink highlighting
- 🟡 **Medium Confidence** (80-89%) - Noticeable red tinting
- 🔴 **High Confidence** (90%+) - Strong ink effect with animation

## Installation

### Development Setup

1. Clone the repository:
   ```bash
   git clone git@github.com:QRY91/slopsquid.git
   cd slopsquid
   ```

2. Load the extension in Chrome:
   - Open Chrome and go to `chrome://extensions/`
   - Enable "Developer mode" (top right)
   - Click "Load unpacked" and select the `slopsquid` directory

3. The SlopSquid icon should appear in your browser toolbar!

### Testing

Visit any website with substantial text content and watch SlopSquid mark suspicious AI-generated sections with ink effects.

## Usage

- **Auto-Detection** - SlopSquid automatically scans pages as they load
- **Manual Toggle** - Click the extension icon to enable/disable
- **Context Menu** - Right-click selected text for instant analysis
- **Tooltips** - Hover over marked text to see confidence scores

## Development

This extension is built with vanilla JavaScript and follows Chrome Extension Manifest V3 standards.

### File Structure
```
slopsquid/
├── manifest.json         # Extension configuration
├── content-script.js     # Main detection logic
├── background.js         # Service worker
├── styles.css           # Ink effect animations
├── popup/               # Extension popup UI
├── options/             # Settings page
└── icons/              # Extension icons
```

### Future Enhancements

- [ ] Machine learning-based detection
- [ ] WebGL shader effects for advanced ink animations
- [ ] User training system for personalized detection
- [ ] Integration with external AI detection APIs
- [ ] Site-specific whitelisting/blacklisting
- [ ] Export detected content for analysis

## Contributing

Part of the [QRY Tool Ecosystem](https://github.com/QRY91) - building privacy-first developer tools.

## License

MIT License - see LICENSE file for details

---

*Inspired by Splatoon's ink mechanics and the need to identify AI-generated content in our daily browsing.* 