import { http, HttpResponse } from 'msw'

export const handlers = [
  http.post('/npan.v1.AppService/GetSearchConfig', () => {
    return HttpResponse.json({
      provider: 'meilisearch',
      host: '',
      indexName: '',
      searchApiKey: '',
      instantsearchEnabled: false,
    })
  }),

  http.post('/npan.v1.AppService/AppSearch', () => {
    return HttpResponse.json({
      result: {
        items: [],
        total: '0',
      },
    })
  }),

  http.post('/npan.v1.AppService/AppDownloadURL', () => {
    return HttpResponse.json({
      result: {
        fileId: '1',
        downloadUrl: 'https://example.com/file.pdf',
      },
    })
  }),
]
