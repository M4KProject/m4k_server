const bcrypt = require('bcrypt');

// Environment variables
const {
    PORT,
    NODERED_EMAIL,
    NODERED_PASSWORD,
    ADMIN_EMAIL,
    ADMIN_PASSWORD,
    S3_BUCKET,
    S3_REGION,
    S3_ENDPOINT,
    S3_ACCESS_KEY,
    S3_SECRET_KEY,
} = process.env;

module.exports = {
    // Basic configuration
    uiPort: PORT || 1880,
    
    // Enable HTTPS if certificates are provided
    https: null,
    
    // User Directory - where Node-RED stores user data
    userDir: '/data/',
    
    // Flow file settings
    flowFile: 'flows.json',
    
    // Function timeout (in seconds)
    functionTimeout: 0,
    
    // Function global context
    functionGlobalContext: {
        // Environment variables available in functions
        ADMIN_EMAIL,
        ADMIN_PASSWORD,
        S3_BUCKET,
        S3_REGION,
        S3_ENDPOINT,
        S3_ACCESS_KEY,
        S3_SECRET_KEY,
    },
    
    // Logging configuration
    logging: {
        console: {
            level: "info",
            metrics: false,
            audit: false
        }
    },
    
    // Editor theme
    editorTheme: {
        projects: {
            enabled: false
        }
    },
    
    // Security configuration
    adminAuth: {
        type: "credentials",
        users: [{
            username: NODERED_EMAIL,
            password: bcrypt.hashSync(NODERED_PASSWORD, 8),
            permissions: "*"
        }]
    },
    
    
    // Static content
    httpStatic: '/data/public/',
    
    // Enable CORS for API
    httpNodeCors: {
        origin: "*",
        methods: "GET,PUT,POST,DELETE"
    },
    
    // Context storage
    contextStorage: {
        default: {
            module: "localfilesystem"
        }
    }
};