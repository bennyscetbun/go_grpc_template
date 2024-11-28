const path = require('path');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');

module.exports = {
  entry: './src/index.ts',
  devtool: 'inline-source-map',
  module: {
    rules:
      [
        {
          test: /\.tsx?$/,
          use: 'ts-loader',
          exclude: /node_modules/,
        },
      ],
  },
  devServer: {
    static: './dist',
    hot: false,
    liveReload: false,
    webSocketServer: false,
    watchFiles: ['./src/**/*', './src/*'],
    setupExitSignals: true,
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js'],
    plugins:
      [
        new TsconfigPathsPlugin({
          configFile: './tsconfig.json',
          extensions: ['.tsx', '.ts', '.js'],
        }),
      ]
  },
  plugins: [
    new CopyWebpackPlugin({
      patterns:
        [
          { from: './assets', to: '' }  // to the dist root directory
        ],
    })
  ],
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist'),
  },
};