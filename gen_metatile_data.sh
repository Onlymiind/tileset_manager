header="SECTION \"ROM Bank \$000\", ROM0[\$0]"

for i in $1/*_metatiles.asm; do
    touch "${i}.temp"
    { echo $header; cat $i; } > "${i}.temp"
    name="${i}_standalone"
    cp "${i}.temp" "${name}.asm"
    rgbasm "./${name}.asm" -o "${name}.o"
    rgblink "${name}.o" -o "${i/_metatiles.asm/".mtile"}"
    rm "${i}.temp"
    rm "${name}.asm"
    rm "${name}.o"
done
