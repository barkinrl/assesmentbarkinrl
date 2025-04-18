/**
 * This work is licensed under Apache License, Version 2.0 or later. 
 * Please read and understand latest version of Licence.
 */
const ReactCompilerConfig = {
  target: '19'
};

module.exports = function () {
  return {
    plugins: [
      ['babel-plugin-react-compiler', ReactCompilerConfig],
    ],
  };
};
