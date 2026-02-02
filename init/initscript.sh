#!/bin/bash
COMMAND="SELECT 1"
chown postgres: -R /init && echo "Change user"
while [ "$CHECK_ROW" != "1" ]
do
        CHECK_ROW=$(su postgres -c "psql -A -X -t -w -c \"${COMMAND}\"")
        sleep 1
done
su postgres -c "psql -d workspace -f /init/init.sql" 2>&1 && echo "Run SQL"
