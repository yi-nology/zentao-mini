/**
 * 外部链接打开(纯网页版:新标签页打开)
 */
export async function openExternalLink(url: string): Promise<void> {
  if (!url) return
  window.open(url, '_blank', 'noopener,noreferrer')
}
