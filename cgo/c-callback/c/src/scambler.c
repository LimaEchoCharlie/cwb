#include <stdio.h>
#include <stdlib.h>
#include <libscramble.h>

// cb is a callback function that swaps "cat" <-> "hat"
Result cb(char* v){
    printf("cb start msg: %s\n", v);
    int i = 0;
    while(v[i] != '\0'){
        if( i > 1 && v[i-1] == 'a' && v[i] == 't'){
            switch(v[i-2]){
            case 'c':
                v[i-2] = 'h';
                break;
            case 'h':
                v[i-2] = 'c';
                break;
            }
        }
        i++;
    }
    printf("cb finish msg: %s\n", v);
    return Success;
}

int main(int argc, char *argv[]) {
    scramble_message(cb);
    printf("Done\n");
    return 0;
}