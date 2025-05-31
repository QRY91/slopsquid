# SlopSquid 🦑

**Detect AI-generated text with beautiful ink effects**

*"Slap that AI!"* - A browser extension that identifies suspicious AI-generated content and marks it with stunning visual feedback as you browse the web.

> **🚧 Early Development** - Core functionality working, polishing in progress before Chrome Web Store release.

## ✨ Features

- 🔍 **Real-time Detection** - Automatically scans text as pages load
- 🎨 **Ink Effects** - Beautiful hot pink to purple visual feedback
- 🎚️ **Adjustable Sensitivity** - Tune detection to your preferences
- 🖱️ **One-Click Toggle** - Easy enable/disable from browser toolbar
- 🔒 **Privacy-First** - All processing happens locally in your browser
- 🎯 **Context Menu** - Right-click any text for instant analysis

## 🎯 Perfect For

- **Journalists** - Verify content authenticity
- **Students** - Identify AI-generated academic content  
- **Content Creators** - Ensure originality in research
- **General Users** - Stay informed about AI content online

## 🚀 Installation

### Development Install (Current)
1. Download or clone this repository
2. Open Chrome and go to `chrome://extensions/`
3. Enable "Developer mode" (top right toggle)
4. Click "Load unpacked" and select the `slopsquid` folder
5. The SlopSquid 🦑 icon appears in your toolbar!

### From Chrome Web Store
*Coming after polish phase...*

## 🎨 How It Works

SlopSquid uses enhanced heuristic analysis to identify common AI writing patterns:

- **Language Analysis** - Detects overly formal or repetitive phrasing
- **Sentence Structure** - Identifies unnaturally consistent patterns
- **AI Vocabulary** - Recognizes common AI-generated phrases
- **Writing Style** - Spots corporate-speak and artificial formality
- **Pattern Matching** - Catches phrases like "large language model", "trained on massive dataset"

### Visual Feedback
- 🟢 **Low Confidence** (60-79%) - Subtle pink highlighting
- 🟡 **Medium Confidence** (80-89%) - Noticeable ink effects  
- 🔴 **High Confidence** (90%+) - Strong ink animation with tooltip

## 🛠️ Usage

1. **Automatic Scanning** - Browse normally, SlopSquid works in the background
2. **Click the Icon** - View detection stats and toggle on/off
3. **Right-click Text** - Analyze specific selections instantly
4. **Open Options** - Fine-tune sensitivity and visual preferences

## 🌐 Website

Visit [slopsquid.com](https://slopsquid.com) for more information, examples, and updates.

## 🔧 Technical Details

Built with modern web technologies:
- **Manifest V3** - Latest Chrome extension standards
- **Vanilla JavaScript** - Fast and lightweight
- **CSS Animations** - Smooth ink effects
- **Local Processing** - No external API dependencies

### File Structure
```
slopsquid/
├── manifest.json         # Extension configuration
├── content-script.js     # Main detection engine
├── background.js         # Service worker
├── styles.css           # Ink effect animations
├── popup/               # Extension popup UI
├── options/             # Settings page
├── icons/              # Crispy pixel art icons
└── index.html          # Landing page
```

## 🎮 Development Status

**✅ Completed:**
- Core AI detection algorithm
- Visual ink effects system
- Chrome extension structure  
- Landing page with ocean/pink theme
- Options and popup interfaces

**🔧 Polish Phase (Next):**
- Sensitivity fine-tuning
- Visual effect refinements
- Custom theming system
- Enhanced detection patterns
- Performance optimizations

**🚀 Future Enhancements:**
- Machine learning integration
- WebGL shader effects
- User training system
- Chrome Web Store release

## 🎮 Inspiration

Inspired by Splatoon's ink mechanics and the growing need to identify AI-generated content in our daily browsing experience.

## 🤝 Contributing

Part of the [QRY Tool Ecosystem](https://github.com/QRY91) - building privacy-first tools for the modern web.

## 📄 License

MIT License - see LICENSE file for details

---

*Made with 🦑 and a commitment to transparency in the age of AI* 