<script lang="ts">
  type Info = { message: string; email: string; env: Record<string,string> }
  let info: Info | null = null
  let error = ''
  async function load() {
    try {
      const res = await fetch('/api/info')
      info = await res.json()
    } catch (e:any) {
      error = e.message
    }
  }
  load()
</script>

<div class="container col">
  <div class="row" style="justify-content: space-between;">
    <div class="col">
      <h2 style="margin:0;">Forge Template App</h2>
      <div class="muted">Svelte frontend + Go backend • Bun + latest Go</div>
    </div>
  </div>

  <div class="panel col">
    <div><b>About</b></div>
    <div>This is a test application. It demonstrates a Forge-compatible Docker Compose, connectors, Cloudflare Access email propagation, and env vars.</div>
    <div class="muted">Design tokens and theme follow the Daedalus guidelines in <code>design_guidelines.md</code>.</div>
  </div>

  {#if error}
    <div class="panel">{error}</div>
  {:else if !info}
    <div class="panel">Loading...</div>
  {:else}
    <div class="panel col">
      <div class="row" style="justify-content: space-between;">
        <div><b>Cloudflare Access email</b></div>
        <div>{info.email}</div>
      </div>
      <div style="margin-top: 8px;"><b>Environment variables</b></div>
      <pre>{Object.entries(info.env).map(([k,v])=>`${k}=${v}`).join('\n')}</pre>
    </div>
  {/if}
</div>


