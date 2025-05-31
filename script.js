// SlopSquid Landing Page JavaScript 🦑

document.addEventListener('DOMContentLoaded', function() {
    
    // Smooth scrolling for navigation links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Navbar background on scroll
    window.addEventListener('scroll', function() {
        const navbar = document.querySelector('.navbar');
        if (window.scrollY > 100) {
            navbar.style.background = 'rgba(15, 23, 42, 0.95)';
        } else {
            navbar.style.background = 'rgba(15, 23, 42, 0.8)';
        }
    });

    // Intersection Observer for fade-in animations
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver(function(entries) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Observe all feature cards and sections
    document.querySelectorAll('.feature-card, .step, .demo-sample').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(30px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });

    // Add ripple effect to primary buttons
    document.querySelectorAll('.btn-primary').forEach(button => {
        button.addEventListener('click', function(e) {
            const ripple = document.createElement('div');
            const rect = this.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;
            
            ripple.style.cssText = `
                position: absolute;
                width: ${size}px;
                height: ${size}px;
                left: ${x}px;
                top: ${y}px;
                background: rgba(255, 255, 255, 0.3);
                border-radius: 50%;
                transform: scale(0);
                animation: ripple 0.6s ease-out;
                pointer-events: none;
            `;
            
            this.appendChild(ripple);
            
            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });

    // Demo samples interactive effects
    document.querySelectorAll('.demo-sample').forEach(sample => {
        sample.addEventListener('mouseenter', function() {
            this.style.transform = 'translateX(10px) scale(1.02)';
        });
        
        sample.addEventListener('mouseleave', function() {
            this.style.transform = 'translateX(0) scale(1)';
        });
    });

    // Animated counter for stats (if we add them later)
    function animateCounter(element, target, duration = 2000) {
        let start = 0;
        const increment = target / (duration / 16);
        
        const timer = setInterval(() => {
            start += increment;
            element.textContent = Math.floor(start);
            
            if (start >= target) {
                element.textContent = target;
                clearInterval(timer);
            }
        }, 16);
    }

    // Parallax effect for hero squid
    window.addEventListener('scroll', function() {
        const scrolled = window.pageYOffset;
        const squidContainer = document.querySelector('.squid-container');
        
        if (squidContainer) {
            const rate = scrolled * -0.5;
            squidContainer.style.transform = `translateY(${rate}px)`;
        }
    });

    // Add floating ink particles effect
    function createInkParticle() {
        const particle = document.createElement('div');
        particle.className = 'ink-particle';
        
        const size = Math.random() * 10 + 5;
        const x = Math.random() * window.innerWidth;
        const duration = Math.random() * 3 + 2;
        
        particle.style.cssText = `
            position: fixed;
            width: ${size}px;
            height: ${size}px;
            background: radial-gradient(circle, #ff1493, #7c3aed);
            border-radius: 50%;
            left: ${x}px;
            top: 100vh;
            opacity: 0.6;
            pointer-events: none;
            z-index: 1;
            animation: floatUp ${duration}s ease-in-out forwards;
        `;
        
        document.body.appendChild(particle);
        
        setTimeout(() => {
            particle.remove();
        }, duration * 1000);
    }

    // Create ink particles periodically
    setInterval(createInkParticle, 3000);

    // Add CSS animation for floating particles
    const style = document.createElement('style');
    style.textContent = `
        @keyframes ripple {
            to {
                transform: scale(2);
                opacity: 0;
            }
        }
        
        @keyframes floatUp {
            0% {
                transform: translateY(0) rotate(0deg);
                opacity: 0.6;
            }
            50% {
                opacity: 0.8;
            }
            100% {
                transform: translateY(-100vh) rotate(360deg);
                opacity: 0;
            }
        }
    `;
    document.head.appendChild(style);

    // Enhanced squid animation on scroll
    let isSquidAnimating = false;
    window.addEventListener('scroll', function() {
        if (!isSquidAnimating) {
            isSquidAnimating = true;
            
            const squidIcon = document.querySelector('.squid-icon');
            if (squidIcon) {
                squidIcon.style.transform = 'scale(1.1) rotate(5deg)';
                
                setTimeout(() => {
                    squidIcon.style.transform = 'scale(1) rotate(0deg)';
                    isSquidAnimating = false;
                }, 300);
            }
        }
    });

    // Add hover effect to nav logo
    const navLogo = document.querySelector('.nav-logo');
    if (navLogo) {
        navLogo.addEventListener('mouseenter', function() {
            this.querySelector('.logo-icon').style.animation = 'pulse 0.5s ease-in-out';
        });
    }

    console.log('🦑 SlopSquid landing page loaded successfully!');
}); 