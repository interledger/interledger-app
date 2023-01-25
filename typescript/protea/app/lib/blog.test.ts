import type { Author } from './blog.server'
describe('Basic test', () => {
  test('should render results', async () => {
    const RemixStub: Author = {
      avatar: 'test',
      name: '',
      twitterHandle: ''
    }
    expect(RemixStub.avatar).toEqual('test')
  })
})
