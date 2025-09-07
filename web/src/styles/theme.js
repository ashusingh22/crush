const theme = {
  colors: {
    // Primary brand colors
    primary: '#6366f1',
    primaryHover: '#4f46e5',
    primaryLight: '#e0e7ff',
    
    // Background colors
    background: '#0f0f0f',
    surface: '#1a1a1a',
    surfaceHover: '#262626',
    
    // Text colors
    text: {
      primary: '#ffffff',
      secondary: '#a3a3a3',
      muted: '#737373',
    },
    
    // UI colors
    border: '#404040',
    borderLight: '#525252',
    success: '#10b981',
    warning: '#f59e0b',
    error: '#ef4444',
    info: '#3b82f6',
    
    // Terminal colors
    terminal: {
      background: '#0a0a0a',
      text: '#00ff00',
      cursor: '#ffffff',
      selection: '#333333',
    },
    
    // Chat colors
    chat: {
      userMessage: '#1e40af',
      assistantMessage: '#059669',
      systemMessage: '#7c3aed',
    }
  },
  
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
    xxl: '3rem',
    xxxl: '4rem',
  },
  
  borderRadius: {
    sm: '0.25rem',
    md: '0.5rem',
    lg: '0.75rem',
    xl: '1rem',
    full: '9999px',
  },
  
  shadows: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  },
  
  breakpoints: {
    mobile: '480px',
    tablet: '768px',
    desktop: '1024px',
    wide: '1280px',
  },
  
  fonts: {
    primary: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
    mono: "'JetBrains Mono', 'Monaco', 'Consolas', 'Menlo', monospace",
  },
  
  animations: {
    fast: '0.1s ease',
    normal: '0.2s ease',
    slow: '0.3s ease',
  },
};

export default theme;