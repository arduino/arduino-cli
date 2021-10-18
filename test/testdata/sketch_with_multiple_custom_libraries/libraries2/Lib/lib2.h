
#if defined __has_include
#  if __has_include (<lib2_user_config.h>)
#    pragma message "Including local project config file <lib2_user_config.h>"
#    include <lib2_user_config.h>
#    define DEFAULT_VALUES_ARE_GIVEN 1
#  endif
#endif

#ifndef LIB2_SOME_CONFIG
#define LIB2_SOME_CONFIG 0
#endif

#define LIB2_SOME_SIZE ((LIB2_SOME_CONFIG) * 42)
