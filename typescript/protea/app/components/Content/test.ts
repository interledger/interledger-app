
let content = {
  id: '122272899',
  title: 'Connect. Verify. Transact with certainty.',
  iterations: [
    {
      id: '122272895',
      title: 'Connect.'
    },
    {
      id: '122272896',
      title: 'Verify.'
    },
    {
      id: '122272897',
      title: 'Transact'
    },
    {
      id: '122272898',
      title: 'certainty.'
    }
  ],
}

type HeroContentRecord = {
  body?: string;
  id: string;
  title?: string;
};

type Segment = {
  text: string
  animated: boolean
}

function getSegments(title: string, iterations: HeroContentRecord[], type: 'title' | 'body'): Segment[] {
  const segments: Segment[] = []
  let lastIndex = 0
  for (const iteration of iterations) {
    const start = title?.indexOf(iteration[type] as string)
    const end = start + (iteration[type] as string).length

    if (start == -1) continue

    if (start > lastIndex) {
      segments.push({
        animated: false,
        text: title.slice(lastIndex, start)
      })
    }

    segments.push({
      animated: true,
      text: iteration[type] as string
    })
    lastIndex = end
  }

  console.log('segments', segments)
  return segments
}

getSegments(content.title, content.iterations , 'title')