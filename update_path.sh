CURRENT_PATH=$1
NEW_JAVA_PATH=$2

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
CURRENT_PATH="$NEW_JAVA_PATH:$CURRENT_PATH"
echo "export JAVA_HOME=\"$NEW_JAVA_PATH\""
echo "export PATH=\"$CURRENT_PATH\""
