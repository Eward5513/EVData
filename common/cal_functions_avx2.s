#include "textflag.h"

// func DistanceAVX2(x1, y1, x2, y2 float64) float64
TEXT ·DistanceAVX2(SB), NOSPLIT, $0-40
    // 加载参数到寄存器
    MOVSD x1+0(FP), X0  // X0 = x1
    MOVSD y1+8(FP), X1  // X1 = y1
    MOVSD x2+16(FP), X2 // X2 = x2
    MOVSD y2+24(FP), X3 // X3 = y2

    // 计算 x2-x1 并存入 X4
    VMOVAPD X2, X4
    VSUBSD  X0, X4, X4  // X4 = x2-x1

    // 计算 y2-y1 并存入 X5
    VMOVAPD X3, X5
    VSUBSD  X1, X5, X5  // X5 = y2-y1

    // 计算 (x2-x1)^2 并存入 X4
    VMULSD X4, X4, X4   // X4 = (x2-x1)^2

    // 计算 (y2-y1)^2 并存入 X5
    VMULSD X5, X5, X5   // X5 = (y2-y1)^2

    // 计算 (x2-x1)^2 + (y2-y1)^2 并存入 X0
    VADDSD X5, X4, X0   // X0 = (x2-x1)^2 + (y2-y1)^2

    // 计算平方根
    VSQRTSD X0, X0, X0  // X0 = sqrt((x2-x1)^2 + (y2-y1)^2)

    // 返回结果
    MOVSD X0, ret+32(FP)
    RET

// func P2lDistanceAVX2(x1, y1, x2, y2, x3, y3 float64) float64
TEXT ·P2lDistanceAVX2(SB), NOSPLIT, $0-56
    // 加载参数到寄存器
    MOVSD x1+0(FP), X0   // X0 = x1
    MOVSD y1+8(FP), X1   // X1 = y1
    MOVSD x2+16(FP), X2  // X2 = x2
    MOVSD y2+24(FP), X3  // X3 = y2
    MOVSD x3+32(FP), X4  // X4 = x3
    MOVSD y3+40(FP), X5  // X5 = y3

    // 计算 (x2-x1)
    VMOVAPD X2, X6
    VSUBSD  X0, X6, X6   // X6 = x2-x1

    // 计算 (y3-y1)
    VMOVAPD X5, X7
    VSUBSD  X1, X7, X7   // X7 = y3-y1

    // 计算 (y2-y1)
    VMOVAPD X3, X8
    VSUBSD  X1, X8, X8   // X8 = y2-y1

    // 计算 (x3-x1)
    VMOVAPD X4, X9
    VSUBSD  X0, X9, X9   // X9 = x3-x1

    // 计算 (x2-x1)*(y3-y1)
    VMULSD X7, X6, X10   // X10 = (x2-x1)*(y3-y1)

    // 计算 (y2-y1)*(x3-x1)
    VMULSD X9, X8, X11   // X11 = (y2-y1)*(x3-x1)

    // 计算 (x2-x1)*(y3-y1) - (y2-y1)*(x3-x1)
    VSUBSD X11, X10, X12 // X12 = (x2-x1)*(y3-y1) - (y2-y1)*(x3-x1)

    // 取绝对值
    VANDPD X12, X12, X12 // 取绝对值

    // 计算分母: Distance(x1, y1, x2, y2)
    // 复用前面的计算结果 X6 = x2-x1, X8 = y2-y1
    VMULSD X6, X6, X13   // X13 = (x2-x1)^2
    VMULSD X8, X8, X14   // X14 = (y2-y1)^2
    VADDSD X14, X13, X13 // X13 = (x2-x1)^2 + (y2-y1)^2
    VSQRTSD X13, X13, X13 // X13 = sqrt((x2-x1)^2 + (y2-y1)^2)

    // 计算最终结果: |numerator| / denominator
    VDIVSD X13, X12, X0  // X0 = |numerator| / denominator

    // 返回结果
    MOVSD X0, ret+48(FP)
    RET

// func CalTAVX2(x1, y1, x2, y2, x3, y3 float64) float64
TEXT ·CalTAVX2(SB), NOSPLIT, $0-56
    // 加载参数到寄存器
    MOVSD x1+0(FP), X0   // X0 = x1
    MOVSD y1+8(FP), X1   // X1 = y1
    MOVSD x2+16(FP), X2  // X2 = x2
    MOVSD y2+24(FP), X3  // X3 = y2
    MOVSD x3+32(FP), X4  // X4 = x3
    MOVSD y3+40(FP), X5  // X5 = y3

    // 计算 (x2-x1)
    VMOVAPD X2, X6
    VSUBSD  X0, X6, X6   // X6 = x2-x1

    // 计算 (x3-x1)
    VMOVAPD X4, X7
    VSUBSD  X0, X7, X7   // X7 = x3-x1

    // 计算 (y2-y1)
    VMOVAPD X3, X8
    VSUBSD  X1, X8, X8   // X8 = y2-y1

    // 计算 (y3-y1)
    VMOVAPD X5, X9
    VSUBSD  X1, X9, X9   // X9 = y3-y1

    // 计算分子: (x2-x1)*(x3-x1) + (y2-y1)*(y3-y1)
    VMULSD X7, X6, X10   // X10 = (x2-x1)*(x3-x1)
    VMULSD X9, X8, X11   // X11 = (y2-y1)*(y3-y1)
    VADDSD X11, X10, X10 // X10 = (x2-x1)*(x3-x1) + (y2-y1)*(y3-y1)

    // 计算分母: (x2-x1)^2 + (y2-y1)^2
    VMULSD X6, X6, X11   // X11 = (x2-x1)^2
    VMULSD X8, X8, X12   // X12 = (y2-y1)^2
    VADDSD X12, X11, X11 // X11 = (x2-x1)^2 + (y2-y1)^2

    // 计算最终结果: 分子 / 分母
    VDIVSD X11, X10, X0  // X0 = 分子 / 分母

    // 返回结果
    MOVSD X0, ret+48(FP)
    RET

// func CalPAVX2(x1, x2, y1, y2, tt float64) (x, y float64)
TEXT ·CalPAVX2(SB), NOSPLIT, $0-48
    // 加载参数到寄存器
    MOVSD x1+0(FP), X0   // X0 = x1
    MOVSD x2+8(FP), X1   // X1 = x2
    MOVSD y1+16(FP), X2  // X2 = y1
    MOVSD y2+24(FP), X3  // X3 = y2
    MOVSD tt+32(FP), X4  // X4 = tt

    // 计算 x2-x1
    VMOVAPD X1, X5
    VSUBSD  X0, X5, X5   // X5 = x2-x1

    // 计算 y2-y1
    VMOVAPD X3, X6
    VSUBSD  X2, X6, X6   // X6 = y2-y1

    // 计算 tt*(x2-x1)
    VMULSD X4, X5, X5    // X5 = tt*(x2-x1)

    // 计算 tt*(y2-y1)
    VMULSD X4, X6, X6    // X6 = tt*(y2-y1)

    // 计算 x1 + tt*(x2-x1)
    VADDSD X0, X5, X0    // X0 = x1 + tt*(x2-x1)

    // 计算 y1 + tt*(y2-y1)
    VADDSD X2, X6, X1    // X1 = y1 + tt*(y2-y1)

    // 返回结果
    MOVSD X0, ret+40(FP) // 返回 x 坐标
    MOVSD X1, ret1+48(FP) // 返回 y 坐标
    RET

// func CalEPAVX2(dis float64) float64
TEXT ·CalEPAVX2(SB), NOSPLIT, $0-16
    // 加载参数 dis 到 X0 寄存器
    MOVSD dis+0(FP), X0
    
    // 计算 dis*dis 并存入 X0
    VMULSD X0, X0, X0
    
    // 加载常量 -3.1167365656361935e-06 到 X1
    MOVSD $-3.1167365656361935e-06, X1
    
    // 计算 (dis*dis) * x 并存入 X0
    VMULSD X1, X0, X0
    
    // 将结果存入返回值
    MOVSD X0, ret+8(FP)
    RET
