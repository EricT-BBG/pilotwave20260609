module.exports = {
  devServer: {
    proxy: 'http://localhost:8080',
    public: 'localhost:8080'
  },
  pages: {
    index: {
      entry: 'src/main.js',
      template: 'public/index.html',
      filename: 'index.html',
      title: 'Pilotwave',
      chunks: ['chunk-vendors', 'chunk-common', 'index']
    },
  },
  "transpileDependencies": [
    "vuetify"
  ],
  publicPath: process.env.NODE_ENV === 'production'
  ? '/dist'
  : '/'
}
