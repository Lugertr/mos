module.exports = {
  '/api/**': {
    target: 'http://localhost:8000/',
    changeOrigin: true,
    secure: false,
  },
  '/auth/**': {
    target: 'http://localhost:8000/',
    changeOrigin: true,
    secure: false,
  }
};
