#include <QGuiApplication>
#include <QDesktopServices>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QTimer>
#include <QUrl>

// Polls /api/v1/desktop/messages with RP_BEARER and opens target_url for open_run / open_workflow.
int main(int argc, char *argv[])
{
    QGuiApplication app(argc, argv);
    QNetworkAccessManager nam;

    const QString base = qEnvironmentVariable("RP_API_BASE", QStringLiteral("http://127.0.0.1:9810"));
    const QString bearer = qEnvironmentVariable("RP_BEARER");

    QTimer health;
    health.setInterval(15000);
    QObject::connect(&health, &QTimer::timeout, [&]() {
        QNetworkRequest req(QUrl(base + QLatin1String("/healthz")));
        nam.get(req);
    });
    health.start();

    if (!bearer.isEmpty()) {
        QTimer poll;
        poll.setInterval(8000);
        QObject::connect(&poll, &QTimer::timeout, [&]() {
            QNetworkRequest req(QUrl(base + QLatin1String("/api/v1/desktop/messages")));
            req.setRawHeader("Authorization", (QStringLiteral("Bearer ") + bearer).toUtf8());
            QNetworkReply *reply = nam.get(req);
            QObject::connect(reply, &QNetworkReply::finished, [reply]() {
                reply->deleteLater();
                if (reply->error() != QNetworkReply::NoError)
                    return;
                const QByteArray body = reply->readAll();
                const QJsonDocument doc = QJsonDocument::fromJson(body);
                if (!doc.isArray())
                    return;
                for (const QJsonValue &v : doc.array()) {
                    if (!v.isObject())
                        continue;
                    const QJsonObject o = v.toObject();
                    const QString mt = o.value(QStringLiteral("message_type")).toString();
                    const QString tu = o.value(QStringLiteral("target_url")).toString();
                    const QString tid = o.value(QStringLiteral("target_workflow_run_id")).toString();
                    if ((mt == QStringLiteral("open_run") || mt == QStringLiteral("open_workflow")) && !tu.isEmpty())
                        QDesktopServices::openUrl(QUrl(tu));
                    else if (mt == QStringLiteral("open_run") && tid.size() > 0) {
                        // Caller can compose URL against web base if needed:
                        QString web = qEnvironmentVariable("RP_WEB_BASE", QStringLiteral("http://127.0.0.1:9811"));
                        QDesktopServices::openUrl(QUrl(web + QStringLiteral("/runs/") + tid));
                    }
                }
            });
        });
        poll.start();
    }

    return app.exec();
}
