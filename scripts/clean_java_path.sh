CURRENT_PATH=$1

function __path_remove(){  
    local P="$2"
    local D=":${P}:";  
    [ "${D/:$1:/:}" != "$D" ] && P="${D/:$1:/:}";  
    P="${P/#:/}";  
    echo "${P/%:/}";  
}  

for java_path in `which -a java`; do
    java_path=`dirname "$java_path"`
    CURRENT_PATH=`__path_remove "$java_path" "$CURRENT_PATH"`
    CURRENT_PATH=`__path_remove "$java_path/" "$CURRENT_PATH"`
done
echo $CURRENT_PATH
