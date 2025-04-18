/**
 * This work is licensed under Apache License, Version 2.0 or later. 
 * Please read and understand latest version of Licence.
 */
const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const HtmlMinimizerPlugin = require("html-minimizer-webpack-plugin");
const CspHtmlWebpackPlugin = require('csp-html-webpack-plugin');
const CompressionPlugin = require("compression-webpack-plugin");

const htmlPlugin = new HtmlWebpackPlugin(
    Object.assign(
        {},
        {
            template: "./src/index.html",
            filename: "./index.html",
            favicon: "./src/favicon.ico"
        }
    )
);

// check env SSO_HOST exists
// if SSO_HOST exists, add it to the cspHtmlWebpackPlugin

const ssoAuthorityUrl = process.env.SSO_AUTHORITY_URL
const ssoClientId = process.env.SSO_CLIENT_ID;

// get host part of SSO_AUTHORITY_URL
const ssoHost = ssoAuthorityUrl ? new URL(ssoAuthorityUrl).host : null;

const ssoEnabled = ssoHost ? (ssoClientId ? true : false) : false;

const def_plugin = new webpack.DefinePlugin({
    'SSO_ENABLED': ssoEnabled,
    'SSO_AUTHORITY_URL': JSON.stringify(ssoAuthorityUrl),
    'SSO_HOST': JSON.stringify(ssoHost),
    'SSO_CLIENT_ID': JSON.stringify(ssoClientId),
});

const ssoHostCsp = ssoHost ? [ssoHost] : [];

const cspHtmlWebpackPlugin = new CspHtmlWebpackPlugin(
    {
        'default-src': "'none'",
        'base-uri': "'self'",
        'script-src': ["'self'"],
        'style-src': ["'self'"],
        'img-src': ["'self'"],
        'font-src': ["'self'"],
        'connect-src': ["'self'"].concat(ssoHostCsp),
        'worker-src': "'self'",
        'frame-src': ["'self'"].concat(ssoHostCsp),
    },
    {
        enabled: true,
    }
);


const miniCssExtractPlugin = new MiniCssExtractPlugin({
    filename: 'static/css/styles.[contenthash].css',
});

const terserPlugin = new TerserPlugin({
    terserOptions: {
        format: {
            comments: false,
        },
        compress: {
            drop_console: false,
        },
    },
    extractComments: false,
    parallel: true,
});

const cssMinimizerPlugin = new CssMinimizerPlugin({
    minimizerOptions: {
        preset: [
            'default',
            {
                discardComments: {removeAll: true},
            },
        ],
    },
});

const htmlMinimizerPlugin = new HtmlMinimizerPlugin({
    minimizerOptions: {
        collapseWhitespace: true,
        removeComments: true,
        removeRedundantAttributes: true,
        removeScriptTypeAttributes: true,
        removeStyleLinkTypeAttributes: true,
        useShortDoctype: true,
    },
});

const compressionPlugin = new CompressionPlugin();

const webpackConfig = {
    mode: 'development',
    entry: {
        app: './src/index.tsx',
    },
    devtool: 'source-map',
    output: {
        filename: ({runtime}) => {
            return 'static/js/[name].[contenthash].js';
        },
        path: path.resolve(__dirname, 'dist'),
        publicPath: '/',
        assetModuleFilename: 'static/images/[hash][ext][query]',
        clean: true,
    },
    plugins: [
        def_plugin,
        htmlPlugin,
        cspHtmlWebpackPlugin,
        miniCssExtractPlugin,
        compressionPlugin,
    ],
    module: {
        rules: [
            {
                test: /\.(ts|tsx)?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.(js|jsx)$/,
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        presets: ['@babel/preset-env', '@babel/preset-react'],
                    },
                },
            },
            {
                test: /\.css$/,
                exclude: /node_modules/,
                use: [MiniCssExtractPlugin.loader, "css-loader"]
            },
            {
                test: /\.(scss|sass)$/,
                exclude: /node_modules/,
                use: [MiniCssExtractPlugin.loader, "css-loader", "sass-loader"]
            },
            {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                exclude: /node_modules/,
                type: "asset/resource",
            },
        ],
    },
    optimization: {
        minimize: true,
        minimizer: [terserPlugin, cssMinimizerPlugin, htmlMinimizerPlugin],
        splitChunks: {
            cacheGroups: {
                styles: {
                    name: 'styles',
                    type: 'css/mini-extract',
                    chunks: 'all',
                    enforce: true,
                },
            },
        },
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
};

module.exports = webpackConfig;
