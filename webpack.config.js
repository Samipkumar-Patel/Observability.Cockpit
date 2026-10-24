const path = require('path');
const CopyWebpackPlugin = require('copy-webpack-plugin');

// Grafana provides these modules at runtime, so they must not be bundled.
const externals = [
  '@grafana/data',
  '@grafana/runtime',
  '@grafana/ui',
  '@grafana/schema',
  'react',
  'react-dom',
  'react-router-dom',
  'rxjs',
  'lodash',
  '@emotion/css',
  '@emotion/react',
];

module.exports = (env, argv) => ({
  mode: argv.mode === 'production' ? 'production' : 'development',
  devtool: argv.mode === 'production' ? 'source-map' : 'eval-source-map',
  target: 'web',
  entry: path.join(__dirname, 'src', 'module.ts'),
  output: {
    path: path.join(__dirname, 'dist'),
    filename: 'module.js',
    library: { type: 'amd' },
    clean: false,
  },
  externals,
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.[tj]sx?$/,
        exclude: /node_modules/,
        use: {
          loader: 'ts-loader',
          options: { transpileOnly: false },
        },
      },
    ],
  },
  plugins: [
    new CopyWebpackPlugin({
      patterns: [{ from: 'plugin.json', to: '.' }],
    }),
  ],
});
