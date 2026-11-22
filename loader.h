/* 由 scripts/generate_go_loader.py 生成；禁止手改。 */
#ifndef KUI_GO_LOADER_H
#define KUI_GO_LOADER_H

#define KUI_C_STATIC 1
#include "native/include/kui_c.h"

#ifdef __cplusplus
extern "C" {
#endif

/* path 为 UTF-8；NULL/空字符串时 Windows 按 DLL 搜索规则加载 kui_c.dll，Linux / macOS 先找可执行文件
 * 所在目录里的 libkui_c.so / libkui_c.dylib，再交给动态链接器的搜索规则。 */
int kui_go_load(const char* path);
int kui_go_is_loaded(void);
const char* kui_go_loader_error(void);

#ifdef __cplusplus
}
#endif
#endif
