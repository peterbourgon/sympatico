#!/usr/bin/env fish

for user in armenia belarus cyprus denmark eritrea finland guyana hungary iceland japan
    curl -Ss -XPOST "http://localhost:8081/auth/signup?user=$user&pass=hunter2"
end
