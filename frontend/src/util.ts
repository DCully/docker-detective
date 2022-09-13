export const rawBytesToReadableBytes = (val: number) => {
    if (val < 1) {
        return '0 B'
    }
    if (val > 1024*1024*1024) {
        val = Math.round(val / 1024 / 1024 / 1024)
        return val + ' GB'
    }
    if (val > 1024*1024) {
        val = Math.round(val / 1024 / 1024)
        return val + ' MB'
    }
    if (val > 1024) {
        val = Math.round(val / 1024)
        return val + ' KB'
    }
    return val + ' B'
}
