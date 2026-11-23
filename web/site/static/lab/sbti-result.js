import { DIM_EXPLANATIONS, DRUNK_TRIGGER_QUESTION_ID, dimensionMeta, dimensionOrder, NORMAL_TYPES, TYPE_LIBRARY, questions, specialQuestions, } from './sbti-data.js';
function sumToLevel(score) {
    if (score <= 3)
        return 'L';
    if (score === 4)
        return 'M';
    return 'H';
}
function levelNum(level) {
    return { L: 1, M: 2, H: 3 }[level];
}
function parsePattern(pattern) {
    return pattern.replace(/-/g, '').split('');
}
export function shuffle(array) {
    const arr = [...array];
    for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        const a = arr[i];
        const b = arr[j];
        arr[i] = b;
        arr[j] = a;
    }
    return arr;
}
export function buildShuffledQuestionList() {
    const shuffledRegular = shuffle([...questions]);
    const insertIndex = Math.floor(Math.random() * shuffledRegular.length) + 1;
    return [
        ...shuffledRegular.slice(0, insertIndex),
        specialQuestions[0],
        ...shuffledRegular.slice(insertIndex),
    ];
}
export function getVisibleQuestions(shuffledQuestions, answers) {
    const visible = [...shuffledQuestions];
    const gateIndex = visible.findIndex(q => q.id === 'drink_gate_q1');
    if (gateIndex !== -1 && answers['drink_gate_q1'] === 3) {
        visible.splice(gateIndex + 1, 0, specialQuestions[1]);
    }
    return visible;
}
function getDrunkTriggered(answers) {
    return answers[DRUNK_TRIGGER_QUESTION_ID] === 2;
}
export function computeSbtiResult(answers) {
    const rawScores = {};
    const levels = {};
    Object.keys(dimensionMeta).forEach(dim => {
        rawScores[dim] = 0;
    });
    questions.forEach(q => {
        const prev = rawScores[q.dim] ?? 0;
        rawScores[q.dim] = prev + Number(answers[q.id] ?? 0);
    });
    Object.entries(rawScores).forEach(([dim, score]) => {
        levels[dim] = sumToLevel(score);
    });
    const userVector = dimensionOrder.map(dim => levelNum(levels[dim]));
    const ranked = NORMAL_TYPES.map(type => {
        const vector = parsePattern(type.pattern).map(ch => levelNum(ch));
        let distance = 0;
        let exact = 0;
        for (let i = 0; i < vector.length; i++) {
            const diff = Math.abs(userVector[i] - vector[i]);
            distance += diff;
            if (diff === 0)
                exact += 1;
        }
        const similarity = Math.max(0, Math.round((1 - distance / 30) * 100));
        const lib = TYPE_LIBRARY[type.code];
        return {
            ...type,
            ...lib,
            distance,
            exact,
            similarity,
        };
    }).sort((a, b) => {
        if (a.distance !== b.distance)
            return a.distance - b.distance;
        if (b.exact !== a.exact)
            return b.exact - a.exact;
        return b.similarity - a.similarity;
    });
    const bestNormal = ranked[0];
    const drunkTriggered = getDrunkTriggered(answers);
    let finalType;
    let modeKicker = '你的主类型';
    let badge = `匹配度 ${bestNormal.similarity}% · 精准命中 ${bestNormal.exact}/15 维`;
    let sub = '维度命中度较高，当前结果可视为你的第一人格画像。';
    let special = false;
    let secondaryType = null;
    if (drunkTriggered) {
        const drunkLib = TYPE_LIBRARY.DRUNK;
        finalType = {
            ...drunkLib,
            distance: 0,
            exact: 0,
            similarity: 100,
        };
        secondaryType = bestNormal;
        modeKicker = '隐藏人格已激活';
        badge = '匹配度 100% · 酒精异常因子已接管';
        sub = '乙醇亲和性过强，系统已直接跳过常规人格审判。';
        special = true;
    }
    else if (bestNormal.similarity < 60) {
        const hhhh = TYPE_LIBRARY.HHHH;
        finalType = {
            ...hhhh,
            distance: bestNormal.distance,
            exact: bestNormal.exact,
            similarity: bestNormal.similarity,
        };
        modeKicker = '系统强制兜底';
        badge = `标准人格库最高匹配仅 ${bestNormal.similarity}%`;
        sub = '标准人格库对你的脑回路集体罢工了，于是系统把你强制分配给了 HHHH。';
        special = true;
    }
    else {
        finalType = bestNormal;
    }
    return {
        rawScores,
        levels,
        ranked,
        bestNormal,
        finalType,
        modeKicker,
        badge,
        sub,
        special,
        secondaryType,
    };
}
export function getDimExplanation(dim, level) {
    const block = DIM_EXPLANATIONS[dim];
    return block[level];
}
