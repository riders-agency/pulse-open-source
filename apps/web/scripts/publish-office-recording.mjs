import fs from 'node:fs/promises';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

// Run PULSE_RECORD_DEMO=1 PULSE_BASE_URL=... pnpm test:e2e e2e/demo.spec.ts first.
const dir = path.resolve('../../artifacts/office-native');
const evidence = JSON.parse(await fs.readFile(path.join(dir, 'evidence.json'), 'utf8'));
const { timings: t } = evidence;
if (evidence.origin !== 'https://pulse.riders.agency') throw new Error('Publish a recording from the main domain only');
if (!t || evidence.artifacts !== 1 || evidence.runCount !== 2 || !evidence.approvedRun || !evidence.rejectedRun) throw new Error('Verified native executions are required');
const video = path.resolve(evidence.video);
if (!video.startsWith(dir + path.sep)) throw new Error('Video must be inside the private test artifact directory');
const clips = [[t.login, 3.5], [t.start, 4], [t.history, 3], [t.approval, 5], [t.confirmed + 0.5, 2.5], [t.results, 4]];
if (clips.some(([start, duration]) => !Number.isFinite(start) || start < 0 || start + duration > t.end + 1)) throw new Error('Invalid recording timeline');
const filters = clips.map(([start, duration], i) => `[0:v]trim=start=${start}:duration=${duration},setpts=PTS-STARTPTS[v${i}]`).join(';') + ';' + clips.map((_, i) => `[v${i}]`).join('') + 'concat=n=6:v=1:a=0[out]';
const media = path.resolve('public/tech/recordings');
const run = args => { const result = spawnSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', ...args], { stdio: 'inherit' }); if (result.status !== 0) throw new Error('Video conversion failed'); };
run(['-i', video, '-filter_complex', filters, '-map', '[out]', '-an', '-c:v', 'libx264', '-preset', 'slow', '-crf', '21', '-pix_fmt', 'yuv420p', '-movflags', '+faststart', path.join(media, 'office.mp4')]);
run(['-i', path.resolve('../../artifacts/tech/office-native-approval.png'), '-frames:v', '1', '-quality', '86', path.join(media, 'office.webp')]);
await fs.writeFile(path.join(media, 'office.vtt'), `WEBVTT

00:00:00.000 --> 00:00:03.500 line:4%
Демо-вход без пароля на pulse.riders.agency.

00:00:03.500 --> 00:00:07.500 line:4%
Своя компания и почтовый сценарий. Сервер проверяет ящик каждые 30 секунд.

00:00:07.500 --> 00:00:10.500 line:4%
История исполнений в приложении. Ожидание опроса сокращено.

00:00:10.500 --> 00:00:15.500 line:4%
ИИ подготовил страницу и счёт. Исполняемый граф ждёт решения человека.

00:00:15.500 --> 00:00:18.000 line:4%
После подтверждения запуск завершён. Показан реальный расход токенов.

00:00:18.000 --> 00:00:22.000 line:4%
Результаты сохранены в демо-компании. Последние 30 исполнений остаются в истории.
`);
console.log(JSON.stringify({ duration_seconds: 22, origin: evidence.origin, mode: 'native-workflow-recording', cuts: 'server wait and navigation omitted' }));
