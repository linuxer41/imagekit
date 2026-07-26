const BASE = 'http://localhost:8080'

async function json(url, opts = {}) {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  const text = await res.text()
  try { return { ok: res.ok, status: res.status, data: JSON.parse(text) } }
  catch { return { ok: res.ok, status: res.status, text } }
}

async function main() {
  let ok = 0, fail = 0

  function check(label, cond, detail = '') {
    if (cond) { ok++; console.log(`  ✅ ${label}`) }
    else { fail++; console.log(`  ❌ ${label} — ${detail}`) }
  }

  // 1. Admin login
  console.log('\n1. Admin login')
  const loginRes = await fetch(BASE + '/admin/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: 'username=zitadel-admin@zitadel.localhost&password=Linuxer2024',
  })
  const loginText = await loginRes.text()
  const tokenMatch = loginText.match(/admin_token','([^']+)/)
  check('login returns token', !!tokenMatch, 'no token in response')
  if (!tokenMatch) { process.exit(1) }
  const adminToken = tokenMatch[1]
  const auth = (t) => ({ Authorization: 'Bearer ' + (t || adminToken) })

  // 2. Admin dashboard stats
  console.log('\n2. Admin stats')
  const stats = await json(BASE + '/admin/api/stats', { headers: auth() })
  check('stats ok', stats.ok, `status ${stats.status}`)

  // 3. Create invitation
  console.log('\n3. Create invitation')
  const inv = await json(BASE + '/api/admin/invitations', {
    method: 'POST', headers: auth(),
  })
  check('invitation created', inv.ok && inv.data.code, inv.ok ? `code: ${inv.data.code}` : `status ${inv.status}`)
  const invCode = inv.ok ? inv.data.code : null

  // 4. List invitations
  console.log('\n4. List invitations')
  const invs = await json(BASE + '/api/admin/invitations', { headers: auth() })
  check('invitations listed', invs.ok && Array.isArray(invs.data), `status ${inv.status}`)

  // 5. List tenants
  console.log('\n5. List tenants')
  const tenants = await json(BASE + '/api/admin/tenants', { headers: auth() })
  check('tenants listed', tenants.ok && Array.isArray(tenants.data), `status ${tenants.status}`)

  // 6. Register new tenant
  console.log('\n6. Register tenant')
  const email = `test-${Date.now()}@example.com`
  const reg = await json(BASE + '/api/panel/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Test User', email, password: 'test123456', invitation_code: invCode }),
  })
  check('tenant registered', reg.ok && reg.data.token, reg.ok ? 'got token' : `status ${reg.status}`)
  const tenantToken = reg.ok ? reg.data.token : null

  // 7. Tenant login
  console.log('\n7. Tenant login')
  const login2 = await json(BASE + '/api/panel/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'test123456' }),
  })
  check('tenant login ok', login2.ok && login2.data.token, `status ${login2.status}`)
  const tToken = (login2.ok ? login2.data.token : null) || tenantToken

  // 8. List projects (empty)
  console.log('\n8. List projects (empty)')
  const projs0 = await json(BASE + '/api/panel/projects', { headers: auth(tToken) })
  check('projects listed', projs0.ok && Array.isArray(projs0.data), `status ${projs0.status}`)

  // 9. Create project via admin API (skips storage validation)
  console.log('\n9. Create project')
  const slug = 'test-' + Date.now()
  const tenantId = login2.ok ? login2.data.tenant_id : 0
  const proj = await json(BASE + '/api/admin/projects', {
    method: 'POST',
    headers: { ...auth(adminToken), 'Content-Type': 'application/json' },
    body: JSON.stringify({ tenant_id: tenantId, slug, name: 'Test Project', provider: 'gcs', bucket: 'test-bucket' }),
  })
  check('project created', proj.ok, proj.ok ? `slug: ${slug}` : JSON.stringify(proj.data))

  // 10. List projects (with the new one)
  console.log('\n10. List projects (with new)')
  const projs1 = await json(BASE + '/api/panel/projects', { headers: auth(tToken) })
  check('project in list', projs1.ok && projs1.data.some(p => p.slug === slug), JSON.stringify(projs1.data))

  // 11. Tenant account
  console.log('\n11. Tenant account')
  const acct = await json(BASE + '/api/panel/account', { headers: auth(tToken) })
  check('account ok', acct.ok && acct.data.email === email, `status ${acct.status}`)

  // 12. Tenant metrics
  console.log('\n12. Tenant metrics')
  const tmetrics = await json(BASE + '/api/panel/metrics?days=7', { headers: auth(tToken) })
  check('tenant metrics ok', tmetrics.ok, `status ${tmetrics.status}`)

  // 13. Admin global metrics
  console.log('\n13. Admin global metrics')
  const ametrics = await json(BASE + '/api/admin/metrics?days=7', { headers: auth() })
  check('admin metrics ok', ametrics.ok, `status ${ametrics.status}`)

  // 14. Admin list users
  console.log('\n14. Admin list users')
  const users = await json(BASE + '/api/admin/users', { headers: auth() })
  check('users listed', users.ok && Array.isArray(users.data), `status ${users.status}`)

  // 15. Project summary
  console.log('\n15. Project summary')
  const projects = projs1.ok ? projs1.data : []
  for (const p of projects) {
    const sum = await json(BASE + `/api/panel/projects/${p.id}/summary`, { headers: auth(tToken) })
    check(`summary for ${p.slug}`, sum.ok, `status ${sum.status}`)
  }

  // 16. List tenants (admin)
  console.log('\n16. Admin list tenants')
  const t2 = await json(BASE + '/api/admin/tenants', { headers: auth() })
  check('tenants listed', t2.ok && Array.isArray(t2.data), `status ${t2.status}`)

  // 17. Health endpoints
  console.log('\n17. Health endpoints')
  const hAdmin = await fetch(BASE + '/health')
  check('admin health ok', hAdmin.ok, `status ${hAdmin.status}`)
  const hImg = await fetch('http://localhost:9000/health')
  check('imager health ok', hImg.ok, `status ${hImg.status}`)

  // Summary
  console.log(`\n${'='.repeat(40)}`)
  console.log(`Result: ${ok} passed, ${fail} failed, ${ok + fail} total`)
  process.exit(fail > 0 ? 1 : 0)
}

main().catch(e => { console.error('FATAL:', e); process.exit(1) })
