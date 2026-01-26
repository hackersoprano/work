#!/bin/bash
chown postgres: -R /init && echo "Change user" 
sleep 5 && echo "Sleped" 
su postgres -c "psql -d workspace -f /init/init.sql" 2>&1 && echo "Run SQL"
