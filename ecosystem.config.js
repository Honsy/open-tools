module.exports = {
    apps : [{
      name:'koassr',
      script: 'dist/app.js',
      env: {
        NODE_ENV: 'dev'
      },
      env_development: {
        NODE_ENV: 'dev'
      },
      env_production: {
        NODE_ENV: 'prod'
      }
    }],
  
    deploy : {
      production : {

      }
    }
  };
  