const path = require('path');
const webpack = require('webpack');
const webpackMerge = require('webpack-merge');

const nodePath = (process.env.NODE_PATH &&
    process.env.NODE_PATH.split(path.delimiter)) || [
    path.join(__dirname, 'node_modules')
];

const modeConfig = env => require(`./webpack-build-utils/webpack.${env}`)(env);
const presetsConfig = require('./webpack-build-utils/loadPresets');


module.exports = ({mode,presets} = {mode:"production",presets:[]})=> {
    console.log('mode,presets',mode,presets)
    const merged = webpackMerge(
        {
            mode,
            module: {
                rules: [
                    {
                        test: /\.jsx?$/,
                        exclude: /node_modules/,
                        use: {
                            loader: 'babel-loader'
                        }
                    },
                    {
                        test: /\.tsx?$/,
                        exclude: /node_modules/,
                        use: [
                            {
                                loader: "ts-loader"
                            }
                        ]
                    },
                    {
                        test: /\.ts?$/,
                        exclude: /node_modules/,
                        use: [
                            {
                                loader: "ts-loader"
                            }
                        ]
                    }
                ]
            },
            resolve: {
                modules: nodePath,
                extensions:['.tsx','.js','.jsx']
            },
            resolveLoader: {
                modules: nodePath
            },
            plugins: [
                new webpack.ProgressPlugin(),
            ]
        },
        modeConfig(mode),
        presetsConfig({ mode, presets })
    );

    return merged;
}
