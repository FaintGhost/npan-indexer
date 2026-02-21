import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v1/app/search', () => {
    return HttpResponse.json({
      items: [],
      total: 0,
    })
  }),

  http.get('/api/v1/app/download-url', () => {
    return HttpResponse.json({
      file_id: 1,
      download_url: 'https://example.com/file.pdf',
    })
  }),
]
