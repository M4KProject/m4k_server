const bcrypt = require('bcrypt');

// Environment variables
const env = process.env;
const PORT = env.PORT || 1880;
const NODERED_ADMIN_EMAIL = env.NODERED_ADMIN_EMAIL;
const NODERED_ADMIN_PASSWORD = env.NODERED_ADMIN_PASSWORD;
const API_ADMIN_EMAIL = env.API_ADMIN_EMAIL;
const API_ADMIN_PASSWORD = env.API_ADMIN_PASSWORD;

module.exports = {
    // Basic configuration
    uiPort: PORT,
    
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
        API_ADMIN_EMAIL: API_ADMIN_EMAIL,
        API_ADMIN_PASSWORD: API_ADMIN_PASSWORD
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
            username: NODERED_ADMIN_EMAIL,
            password: bcrypt.hashSync(NODERED_ADMIN_PASSWORD, 8),
            permissions: "*"
        }]
    },
    
    // HTTP node security
    httpNodeAuth: {
        user: NODERED_ADMIN_EMAIL,
        pass: bcrypt.hashSync(NODERED_ADMIN_PASSWORD, 8)
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