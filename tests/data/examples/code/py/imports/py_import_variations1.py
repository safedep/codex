import math
from operator import itemgetter
from django.core.paginator import Paginator, PageNotAnInteger, EmptyPage
from django.db.models import F, Count, Func
from django.db import connection
from pyclickup import ClickUp
import time
from tsint_service.model_filters import filter_domain

import myIt
from config import URL, KEY, CLICKUP_BASEURL, CLICKUP_LIST_ID, CLICKUP_ACCESS_TOKEN, CLICKUP_TEAM_ID, CLICKUP_SPACE_ID, \
    CLICKUP_FOLDER_ID
import requests
from jinjasql import JinjaSql
from django.db.models import Q, F, Sum
from datetime import datetime, timedelta, timezone, date
from logging_utils import LoggingUtils
from myIt.parser_key import subjectType, ModuleType, TimeLineEntityType
from myIt.utils import getScoreFromSSLGrade, getGradeFromScoreForSsl, dictfetchall_read, get_severity_clause, \
    open_database_tags
from .services import (getSSLDetailsForDomain, getWebScoreRec)
from domainOsint.models import Domain
from .helper.get_score import (
    get_score_from_grade, get_dns_details, get_SSL_Scoring, get_web_grade_from_score)
from myIt.serializers import (
    ApplicationCVELocationSerializer, DomainSerializer, IpSerializer, RelDomSerializer,
    HostSerializer, LeakedCredentialsSerializer, PhishingDetailsSerializer,
    TechnologySerializer, WhoisSerializer, DigitalRiskSerializer, ScoresSerializer, ApiDiscoverySerializer,
    ApiDiscoveryReadSerializer)
from myIt.utils import (
    getDefs, getIntegerDefs, getBooleans, getBoolean, getStructureForRDUpdates, getStructureForHostUpdate,
    getStructureForIPUpdates, dictfetchall, datetime2isostr, compareDateTime, isMainDomain, parse_time,
    uploadFile_bucket,
    get_bucket_url, convert_timestamp, getCurrentDatetime, jsonFromOrderedDict, databasePorts, networkServices,
    preprodTags
)
from .myIt.models import (
    ApplicationCVE, ApplicationCVELocation, DomainStatusType, Ip, RelatedDomain, Host, ScopeType, Scores, Application, LeakedCredentials, PhishingNormDetails, ServiceCVE,
    Technology, Whois, DigitalRisk, Service, ApiDiscovery, PasteData, NetBlock, CVEState, ScopeStatusSend,
    NetbloclOwnershipType, ObservationsAndInsights, ServiceStateSend
)

from fc_cloud_storage_client.main import GlobalCloudStorage
from timeline.timeline_handler import TimelineHandler
from django.db.models.query import QuerySet as querySet
from tsint.settings import SEND_DATA_TO_DW_SERVICE
# from service.common_functions import getss
from async_operation.utils import get_boolean
from django.contrib.postgres.aggregates import ArrayAgg


class PubSubPublisher(object):

    def __init__(self):
        self.publisher = pubsub_v1.PublisherClient()

    def publish(self, topic, data):
        if topic:
            topic_path = self.publisher.topic_path(topic.project, topic.name)
            self.publisher.publish(topic_path, data=data)

