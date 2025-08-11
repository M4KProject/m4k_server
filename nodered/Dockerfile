FROM nodered/node-red:latest

# Install bcrypt for password hashing
USER root
RUN npm install -g bcrypt
USER node-red

# Copy settings file
COPY settings.js /data/settings.js

EXPOSE 1880