// Enable this for debugging
const DEBUG = false; // Disable debug logs in production

export const debugLog = (...args: any[]) => {
  if (DEBUG) {
    console.log('[Debug]', ...args);
  }
}; 