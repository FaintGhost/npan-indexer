import { describe, expect, it } from 'vitest'
import { createSearchStateMapping } from './instantsearch-routing'

describe('createSearchStateMapping', () => {
  it('maps query, page and file_category route state into InstantSearch uiState', () => {
    const mapping = createSearchStateMapping('npan-public')

    expect(mapping.routeToState({
      query: 'report',
      page: '2',
      file_category: 'doc',
    })).toEqual({
      'npan-public': {
        query: 'report',
        page: 2,
        refinementList: {
          file_category: ['doc'],
        },
      },
    })
  })

  it('falls back to the default refinement when file_category route value is invalid', () => {
    const mapping = createSearchStateMapping('npan-public')

    expect(mapping.routeToState({
      query: 'report',
      page: '2',
      file_category: 'invalid',
    })).toEqual({
      'npan-public': {
        query: 'report',
        page: 2,
      },
    })
  })

  it('maps InstantSearch uiState back into shareable query parameters', () => {
    const mapping = createSearchStateMapping('npan-public')

    expect(mapping.stateToRoute({
      'npan-public': {
        query: 'report',
        page: 2,
        refinementList: {
          file_category: ['doc'],
        },
      },
    })).toEqual({
      query: 'report',
      page: 2,
      file_category: 'doc',
    })
  })

  it('omits default page and all-filter state from the route', () => {
    const mapping = createSearchStateMapping('npan-public')

    expect(mapping.stateToRoute({
      'npan-public': {
        query: 'report',
        page: 1,
      },
    })).toEqual({
      query: 'report',
    })
  })
})
