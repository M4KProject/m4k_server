module.exports = {
    uiPort: process.env.PORT || 1880,
    userDir: '/data/',
    flowFile: 'flows.json',
    functionTimeout: 0,
    
    functionGlobalContext: {
        ADMIN_EMAIL: process.env.ADMIN_EMAIL,
        ADMIN_PASSWORD: process.env.ADMIN_PASSWORD
    },
    
    logging: {
        console: {
            level: "info",
            metrics: false,
            audit: false
        }
    },
    
    editorTheme: {
        projects: {
            enabled: false
        }
    },
    
    adminAuth: {
        type: "credentials",
        users: [{
            username: process.env.NODERED_EMAIL || "admin",
            password: process.env.NODERED_PASSWORD || "changeme",
            permissions: "*"
        }]
    },
    
    contextStorage: {
        default: {
            module: "localfilesystem"
        }
    }
};