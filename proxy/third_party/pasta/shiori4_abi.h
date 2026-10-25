/*
 * shiori4_abi.h — SHIORI 次世代 ABI 定義（Rust 版 shiori4_abi と同一レイアウト）
 * 教科書: https://github.com/ekicyou/shiori4/blob/main/doc/textbook/README.md
 *
 * 版（abi_version）
 * 1. SHIORI_ABI_VERSION（4）: 既定。host 必須。ペイロードはヘッダ部分だけ（開始行なし）
 *    - SHIORI は copy_from_slice があるかを確かめ、無ければ load で NULL を返す
 * 2. SHIORI_ABI_VERSION_3（3）: SHIORI3 互換モード。異なる OS の互換ベースウェア向け
 *    - load の host は copy_from_slice だけを持つテーブル（abi_version は 3。request・post・log は NULL）。
 *      SHIORI は copy_from_slice があるかだけを確かめる。ほかのホストの呼び出しは使えない
 *    - request の req は SHIORI/3.0 の要求全文（"GET SHIORI/3.0" / "NOTIFY SHIORI/3.0" で始まる）
 *    - request の out は SHIORI/3.0 の応答全文（"SHIORI/3.0 200 OK" で始まり空行で終わる）
 *    - 戻り値は開始行と同じステータスコード（正本は文字列）。負の値は ABI エラー（out は空。500 として扱う）
 *    - notify の req は "NOTIFY SHIORI/3.0" で始まる全文。GET なら 400
 *    - 文字コードは UTF-8 固定（Charset: UTF-8）。SHIORI/2.x の要求には 400 を返す
 *    - v3 は v4 と同じ機能を出すことを保証しない
 * 3. v4 の文字コードは UTF-8 固定。ホストは要求に Charset ヘッダを付けなくてよい。付いていても
 *    SHIORI は無視する。SHIORI は応答に Charset を付けない
 * 4. ホストは shiori_get_api(v) で版を選ぶ。返るテーブルの abi_version は v と一致する
 *
 * ABI 規約
 * 1. shiori4 の書き出しは shiori_get_api の 1 つだけ。対応できない版なら NULL を返す
 *    （SHIORI/3.0 の書き出しと並べてよい。「DLL の口と発見の順」を参照）
 * 2. DLL 境界を越えるのは C のレイアウトの構造体と関数ポインタだけ
 * 3. ペイロードは UTF-8 のバイト列（ptr と len）。NUL 終端にはしない
 * 4. 引数の (ptr, len) は呼び出しの間だけ有効な借用。保持する場合は受け取った側がコピーする
 * 5. 所有権が移るのは out に書き込まれる ShioriBytes だけ。out の ShioriBytes は常にホストが作り、
 *    ホストの release で解放する（SHIORI は応答を host->copy_from_slice で写して作る）。
 *    写せなかったときは、ホストが自分で決めた印を out に残し（SHIORI には知らせない）、
 *    SHIORI は out をそのまま返す
 * 6. 呼び出された側は out を最後に書く。呼び出した側は戻り値に関係なく out を解放する
 * 7. 構造体は末尾にだけフィールドを追加する（追加する関数ポインタは NULL 許容）。
 *    受け取った側は struct_size を見て、フィールドがあるかを判定する
 * 8. エクスポートする関数は、例外・パニックを境界の外へ出さず SHIORI_E_PANIC を返す
 *
 * 戻り値（ShioriStatus）
 * 1. 0 以上: SHIORI/3.0 のステータスコード（200, 204, 400, 500 など）
 * 2. 負の値: ABI 層のエラー（SHIORI_E_*）。out は空。SHIORI_E_PANIC ならホストはインスタンスを unload する
 *
 * スレッド規約
 * 1. ShioriApi の load / request / notify / unload は、ホストが決めた単一のスレッドからだけ呼ぶ。
 *    このスレッドはインスタンスごとに決まり、プロセスで 1 つではない。
 *    同じ DLL の 2 つのインスタンスは、別々のスレッドから呼ばれうる
 * 2. SHIORI が host->request を呼べるのは、load / request / notify を呼ばれているスレッドから、
 *    その呼び出しが返るまでの間だけ（v4）。
 *    ホストは、これに反する呼び出しを見分けたら SHIORI_E_WRONG_THREAD を返してよい（out は空）。
 *    判定はインスタンスごとに行う
 * 3. host->request の処理中に、ホストから SHIORI を呼び返してはならない（SHIORI_E_REENTRANT）
 * 4. host->post と host->log は任意のスレッドから呼べる。ホストは msg をコピーしてすぐ返る
 *    （post はキューに積む）
 * 5. ShioriHostApi は unload が返るまで有効。SHIORI は、unload が返った後に ShioriHostApi の
 *    どの関数も呼ばない（例外は ShioriBytes.release だけ。規約 6）。
 *    unload を始めたホストは post に SHIORI_E_SHUTDOWN を返してよい。
 *    unload は長く止まらないこと。ホストは OS の終了の猶予の中で、各インスタンスの unload を
 *    呼ぶことがある
 * 6. ShioriBytes.release は常にホストの関数で、いつでも・どのスレッドから呼んでもよい
 *    （インスタンスの unload の後でもよい）。host->copy_from_slice も、どのスレッドから呼んでもよい。
 *    ホストはそれに耐えること
 * 7. notify を即 204 で返し、処理を後で行う実装は合法である。後回しにした処理の中では
 *    host->request を使えない（規約 2 の期間の外であるため）。ホストへの依頼には post を使う
 *
 * load の規約
 * 1. dir は、v4・v3 互換モードとも、絶対パス・OS の区切り・末尾に区切りあり・UTF-8 の文字列
 *    （呼び出し中だけ有効）
 * 2. カレントディレクトリはホスト次第である。SHIORI はカレントディレクトリに頼らず、load の dir を使う
 * 3. v4 の load が失敗するときは、NULL を返す前に host->log で理由を出す。v3 は host に log が無いので、
 *    SHIORI 自身のログに出す。v3 の load は失敗の理由を返さない（NULL を返すだけ）
 * 4. SHIORI は同じプロセスでの 2 つ目のインスタンスを拒否してよい（NULL ＋ log）。拒否されたホストは、
 *    そのゴーストを読めなかったものとして扱い、ゴーストごとに別のプロセスで読むか、拒否を受け入れる
 * 5. SHIORI は、できれば複数のインスタンスを受けるべきである（状態をインスタンスごとに持ち、
 *    グローバルな状態を避ける）
 *
 * ホストのプロセス
 * 1. ホストは SHIORI を専用のプロセスで読むことがある。SHIORI は、ホストと同じプロセスにいること
 *    （ホストの窓・プロセスの状態を共有すること）を前提にしない
 * 2. host->request は IPC を往復することがあり、遅くなりうる
 *
 * DLL の口と発見の順
 * 1. 1 つの DLL が、SHIORI/3.0 の書き出し（load・request・unload）と shiori_get_api の両方を持ってよい
 * 2. ホストは書き出しを shiori_get_api → loadu → load の順に探し、最初に見つかった口で解決する。
 *    shiori_get_api が見つかったら 4、だめなら 3 を試す。どちらも NULL なら読めなかったものとして扱い、
 *    loadu・load へは進まない。解決した口の load が失敗しても、ほかの口へは進まない
 *
 * post の完了通知（v4）
 * 1. 受付 ID は SHIORI が発行し、Tag ヘッダとしてメッセージに入れる（登録してから post する）
 * 2. post の戻り値は受付可否だけ（204: 受付、400: 不正、SHIORI_E_SHUTDOWN: 終了中）
 * 3. 完了はホストが通常のイベント（request または notify）として、post したインスタンスに届ける
 *    - post の呼び出し中や host->request の処理中には届けない（イベントループから届ける）
 *    - unload を始めたら、そのインスタンス宛ての未配送イベントは破棄する
 * 4. Tag を付けない post は「結果は不要」を意味する
 * 5. 運ぶ中身の語（Command・Tag・OnHostComplete の並び・拒否の返し方など）は、shiori4 リポジトリの doc/host-protocol.md（ホスト側プロトコル仕様）が定める
 *
 * 配布の決まり（DLL の選び方）
 * 1. descript.txt の shiori,<名前> から名前の幹を取る（例 shiori,pasta.dll → pasta）
 * 2. ホストは、ホストの OS と arch ごとに次の順で探す
 *    - Windows x64: <幹>.x64.dll → <幹>.dll
 *    - Windows arm64: <幹>.arm64.dll →（x64 のプロセスを起こせるホストだけ）<幹>.x64.dll → <幹>.dll
 *    - Linux x64: lib<幹>.x64.so → lib<幹>.so
 *    - Linux arm64: lib<幹>.arm64.so → lib<幹>.so
 *    - macOS: lib<幹>.dylib（ユニバーサルバイナリを推奨）
 * 3. 接尾辞は x64・arm64 の 2 つだけ。x86 は作らない。接尾辞なしの Windows の DLL が、
 *    32bit の SHIORI/3.0 を担う
 * 4. Linux と macOS は lib の付いた名前だけを探す。macOS は arch の接尾辞を持たない
 * 5. ホストは選んだファイルの機械の種類をヘッダ（PE・ELF・Mach-O）で確かめ、合わなければ次の候補へ
 *    進み、候補が尽きたらそのゴーストを読めなかったものとして扱う。これは SHIORI/3.0 と v4 の両方に使える
 */
#ifndef SHIORI4_ABI_H
#define SHIORI4_ABI_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN32)
#  define SHIORI_EXPORT __declspec(dllexport)
#else
#  define SHIORI_EXPORT __attribute__((visibility("default")))
#endif

/* ============ 定数 ============ */
#define SHIORI_ABI_VERSION      4u   /* 既定 */
#define SHIORI_ABI_VERSION_3    3u   /* SHIORI3 互換モード（host は copy_from_slice だけ） */
#define SHIORI_ABI_VERSION_MIN  3u
#define SHIORI_GET_API_SYMBOL   "shiori_get_api"

typedef int32_t ShioriStatus;

/* SHIORI/3.0 のステータスコード（0 以上） */
#define SHIORI_200_OK                     200
#define SHIORI_204_NO_CONTENT             204
#define SHIORI_311_NOT_ENOUGH             311  /* 互換用。新規実装では非推奨 */
#define SHIORI_312_ADVICE                 312  /* 互換用。新規実装では非推奨 */
#define SHIORI_400_BAD_REQUEST            400
#define SHIORI_500_INTERNAL_SERVER_ERROR  500

/* ABI 層のエラー（負の値） */
#define SHIORI_E_INVALID_ARG   (-1)
#define SHIORI_E_UNSUPPORTED   (-2)
#define SHIORI_E_PANIC         (-3)  /* ホストはこのインスタンスを unload すべき */
#define SHIORI_E_REENTRANT     (-4)
#define SHIORI_E_SHUTDOWN      (-5)
#define SHIORI_E_INVALID_UTF8  (-6)
#define SHIORI_E_WRONG_THREAD  (-7)  /* 規約 2 に反する host->request の呼び出し */

#define SHIORI_IS_SUCCESS(s)    ((s) >= 200 && (s) < 300)
#define SHIORI_IS_ABI_ERROR(s)  ((s) < 0)

/* ログレベル */
#define SHIORI_LOG_ERROR 1
#define SHIORI_LOG_WARN  2
#define SHIORI_LOG_INFO  3
#define SHIORI_LOG_DEBUG 4

/* ============ 所有権付きバイト列 ============
 * ptr/len : UTF-8 のバイト列（NUL 終端ではない）
 * owner   : 確保した側だけが解釈する不透明なポインタ
 * release : 確保した側の解放関数。受け取った側が使い終わったら必ず呼ぶ（NULL なら不要）
 *           release は呼び出し後に *bytes を空（全フィールド 0）にする
 */
typedef struct ShioriBytes {
    const uint8_t* ptr;
    size_t         len;
    void*          owner;
    void         (*release)(struct ShioriBytes* bytes);
} ShioriBytes;

#define SHIORI_BYTES_EMPTY { NULL, 0, NULL, NULL }

static inline void shiori_bytes_release(ShioriBytes* b) {
    if (b && b->release) b->release(b);
}

/* ============ ホスト → SHIORI に渡すテーブル（v3 では copy_from_slice だけ） ============ */
typedef struct ShioriHostApi {
    uint32_t struct_size;
    uint32_t abi_version;   /* 4 以上 */
    void*    ctx;

    /* 同期リクエスト。load / request / notify を呼ばれているスレッドから、その呼び出しが
     * 返るまでの間だけ呼べる（スレッド規約 2）。外れた呼び出しには SHIORI_E_WRONG_THREAD が返りうる
     * 戻り値はステータスコード。out にはヘッダ部分。どの戻り値でも out は解放すること
     * v3 のテーブルでは NULL */
    ShioriStatus (*request)(void* ctx, const uint8_t* req, size_t len, ShioriBytes* out);

    /* 非同期通知。任意のスレッドから呼べる（スレッド規約 4）
     * msg は呼び出し中のみ有効。ホストは保持する場合コピーする
     * 戻り値: 204 受付 / 400 不正 / SHIORI_E_SHUTDOWN 終了中。v3 のテーブルでは NULL */
    ShioriStatus (*post)(void* ctx, const uint8_t* msg, size_t len);

    /* 任意のスレッドから呼べる（スレッド規約 4）。msg は呼び出し中のみ有効。v3 のテーブルでは NULL */
    void         (*log)(void* ctx, int32_t level, const uint8_t* msg, size_t len);

    /* SHIORI の応答を作る。ptr から len バイトをホストのメモリへ写し、ホストが作った ShioriBytes を
     * out に入れる（ABI 規約 5）。任意のスレッドから呼べる（スレッド規約 6）。v3 のテーブルにもある
     * 戻り値は無い。写せなかったときは、ホストが自分で決めた印を out に残す（SHIORI には知らせない）
     * SHIORI は SHIORI_HAS_FIELD(host, ShioriHostApi, copy_from_slice) で有無を確かめる */
    void         (*copy_from_slice)(void* ctx, const uint8_t* ptr, size_t len, ShioriBytes* out);

    /* 将来のフィールドはここに追加する（NULL 許容） */
} ShioriHostApi;

/* ============ SHIORI → ホスト に返すテーブル ============ */
typedef struct ShioriApi {
    uint32_t struct_size;
    uint32_t abi_version;   /* shiori_get_api で要求した版と同じ値 */

    /* インスタンスを生成する。失敗したら NULL（v4 は先に host->log で理由を出す。load の規約）
     * host: v4 では必須（unload が返るまで有効）。v3 では copy_from_slice だけを持つテーブル
     * dir : 絶対パス・OS の区切り・末尾に区切りあり・UTF-8 のゴーストディレクトリ
     *       （呼び出し中のみ有効。load の規約） */
    void*        (*load)(const ShioriHostApi* host /* nullable */,
                         const uint8_t* dir, size_t dir_len);

    void         (*unload)(void* inst);

    /* GET に相当する（v3 では NOTIFY 全文も受け付ける）。戻り値はステータスコード
     * out の形式は版による。どの戻り値でも out は解放すること */
    ShioriStatus (*request)(void* inst, const uint8_t* req, size_t len, ShioriBytes* out);

    /* NOTIFY に相当する。本文は返さない */
    ShioriStatus (*notify)(void* inst, const uint8_t* req, size_t len);

    /* 将来のフィールドはここに追加する（NULL 許容） */
} ShioriApi;

typedef const ShioriApi* (*ShioriGetApiFn)(uint32_t requested_abi_version);

/* shiori4 の唯一のエクスポート（ABI 規約）。要求された版に対応できなければ NULL を返す
 * SHIORI 側でビルドするときだけ SHIORI_BUILD_MODULE を定義する */
#ifdef SHIORI_BUILD_MODULE
SHIORI_EXPORT const ShioriApi* shiori_get_api(uint32_t requested_abi_version);
#endif

/* 追加フィールドが存在するかの判定
 * 例: SHIORI_HAS_FIELD(api, ShioriApi, notify) */
#define SHIORI_HAS_FIELD(p, T, f) \
    ((size_t)(p)->struct_size >= offsetof(T, f) + sizeof(((T*)0)->f))

/* 基本フィールドが揃っているか */
#define SHIORI_API_IS_VALID(p) \
    ((p)->abi_version >= SHIORI_ABI_VERSION_MIN && SHIORI_HAS_FIELD(p, ShioriApi, notify))
#define SHIORI_HOST_API_IS_VALID(p) \
    ((p)->abi_version >= SHIORI_ABI_VERSION && SHIORI_HAS_FIELD(p, ShioriHostApi, log))

/* ============ レイアウト検査（Rust 版と一致すること） ============ */
#if defined(__cplusplus)
#  define SHIORI_STATIC_ASSERT(c, m) static_assert(c, m)
#else
#  define SHIORI_STATIC_ASSERT(c, m) _Static_assert(c, m)
#endif

SHIORI_STATIC_ASSERT(sizeof(ShioriBytes) == 4 * sizeof(void*), "ShioriBytes size");
SHIORI_STATIC_ASSERT(offsetof(ShioriHostApi, ctx) == 8, "ShioriHostApi.ctx offset");
SHIORI_STATIC_ASSERT(offsetof(ShioriHostApi, log) == 8 + 3 * sizeof(void*), "ShioriHostApi.log offset");
SHIORI_STATIC_ASSERT(offsetof(ShioriHostApi, copy_from_slice) == 8 + 4 * sizeof(void*), "ShioriHostApi.copy_from_slice offset");
SHIORI_STATIC_ASSERT(sizeof(ShioriHostApi) == 8 + 5 * sizeof(void*), "ShioriHostApi size");
SHIORI_STATIC_ASSERT(offsetof(ShioriApi, load) == 8, "ShioriApi.load offset");
SHIORI_STATIC_ASSERT(offsetof(ShioriApi, notify) == 8 + 3 * sizeof(void*), "ShioriApi.notify offset");
SHIORI_STATIC_ASSERT(sizeof(ShioriApi) == 8 + 4 * sizeof(void*), "ShioriApi size");

#ifdef __cplusplus
}
#endif
#endif /* SHIORI4_ABI_H */
