cat $1 | grep -o '\$[8-9a-f][0-9a-f]' | sort | uniq -c | sort -bgr > $2
cat $2 | wc -l >> $2