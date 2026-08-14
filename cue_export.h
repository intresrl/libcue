#ifndef CUE_EXPORT_H
#define CUE_EXPORT_H

#if defined(_WIN32)
    #define CUE_EXPORT __declspec(dllexport)
#elif defined(__GNUC__) || defined(__clang__)
    #define CUE_EXPORT __attribute__((visibility("default")))
#else
    #define CUE_EXPORT
#endif

#endif /* CUE_EXPORT_H */
