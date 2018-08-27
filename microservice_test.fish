#!/usr/bin/env fish


curl -Ss -XPOST "localhost:8081/auth/signup?user=charlie&pass=abc123"

echo -n "user=charlie count, want 1, have "
echo "SELECT Count(*) FROM credentials WHERE user = 'charlie';" | sqlite3 auth.db

curl -Ss -XPOST "localhost:8081/auth/login?user=charlie&pass=abc123" | read token

echo -n "after login, token count for charlie, want 1, have "
echo "SELECT Count(*) FROM tokens WHERE user = 'charlie';" | sqlite3 auth.db

set sequence (dna)
set subsequence (echo $sequence | cut -c5-10)
echo sequence $sequence, subsequence $subsequence

curl -Ss -XPOST "localhost:8080/dna/add?user=charlie&token=$token&sequence=$sequence"

echo "SELECT sequence FROM dna WHERE user = 'charlie';" | sqlite3 dna.db | read selected
if test "$selected" = "$sequence" ; echo "sequence check pass" ; else ; echo "sequence check FAIL" ; end

curl -Ss -XGET "localhost:8080/dna/check?user=charlie&token=$token&subsequence=$subsequence"

curl -Ss -XPOST "localhost:8081/auth/logout?user=charlie&token=$token"

echo -n "after logout, token count for charlie, want 0, have "
echo "SELECT Count(*) FROM tokens WHERE user = 'charlie';" | sqlite3 auth.db

