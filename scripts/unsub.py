#!/usr/bin/env python3

import os
import sys
import pickle
import time
import json
from datetime import datetime, timedelta

try:
    import pytz
except ImportError:
    print("Error: 'pytz' library not found. Please install it using 'pip install pytz'")
    sys.exit(1)

from googleapiclient.discovery import build
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request
from googleapiclient.errors import HttpError

# This scope allows for modification of YouTube account details.
SCOPES = ['https://www.googleapis.com/auth/youtube']
API_SERVICE_NAME = 'youtube'
API_VERSION = 'v3'
CLIENT_SECRETS_FILE = 'client_secret.json'
TOKEN_FILE = 'token.pickle'

def authenticate_youtube():
    """Authenticate with YouTube API and return service object."""
    creds = None
    
    if os.path.exists(TOKEN_FILE):
        with open(TOKEN_FILE, 'rb') as token:
            creds = pickle.load(token)
    
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            if not os.path.exists(CLIENT_SECRETS_FILE):
                print(f"Error: {CLIENT_SECRETS_FILE} not found. Please download from Google Cloud Console.")
                sys.exit(1)
            
            flow = InstalledAppFlow.from_client_secrets_file(
                CLIENT_SECRETS_FILE, SCOPES)
            creds = flow.run_local_server(port=0)
        
        with open(TOKEN_FILE, 'wb') as token:
            pickle.dump(creds, token)
    
    return build(API_SERVICE_NAME, API_VERSION, credentials=creds)

def wait_for_quota_reset():
    """Waits until the YouTube API quota resets at midnight Pacific Time."""
    print("\nQuota exceeded. Waiting for reset at midnight Pacific Time...")
    
    pacific = pytz.timezone('America/Los_Angeles')
    now_pacific = datetime.now(pacific)
    reset_time = (now_pacific + timedelta(days=1)).replace(hour=0, minute=0, second=0, microsecond=0)
    wait_seconds = (reset_time - now_pacific).total_seconds()
    
    print(f"Waiting for {int(wait_seconds // 3600)} hours, {int((wait_seconds % 3600) // 60)} minutes.")
    
    while wait_seconds > 0:
        mins, secs = divmod(wait_seconds, 60)
        hours, mins = divmod(mins, 60)
        timer = f"Time until quota reset: {int(hours):02d}:{int(mins):02d}:{int(secs):02d}"
        print(timer, end="\r")
        time.sleep(1)
        wait_seconds -= 1
    
    print("\nResuming...                                       ")

def list_subscriptions(youtube):
    """List all subscriptions for the authenticated user, handling quota errors."""
    subscriptions = []
    next_page_token = None
    
    while True:
        try:
            request = youtube.subscriptions().list(
                part='snippet,id',
                mine=True,
                maxResults=50,
                pageToken=next_page_token
            )
            response = request.execute()
            
            subscriptions.extend(response.get('items', []))
            
            next_page_token = response.get('nextPageToken')
            if not next_page_token:
                break
        except HttpError as e:
            is_quota_error = False
            if e.resp.status == 403:
                try:
                    error_details = json.loads(e.content.decode('utf-8'))
                    if 'error' in error_details and 'errors' in error_details['error']:
                        for error in error_details['error']['errors']:
                            if error.get('reason') == 'quotaExceeded':
                                is_quota_error = True
                                break
                except (json.JSONDecodeError, KeyError):
                    pass
            
            if is_quota_error:
                wait_for_quota_reset()
                continue
            else:
                print(f"An HTTP error {e.resp.status} occurred while listing subscriptions: {e.content}")
                break
            
    return subscriptions

def unsubscribe_from_all(youtube, subscriptions):
    """Unsubscribe from all channels in the list, handling quota errors."""
    total = len(subscriptions)
    i = 0
    while i < total:
        item = subscriptions[i]
        sub_id = item['id']
        channel_title = item['snippet']['title']
        print(f"Unsubscribing from {channel_title} ({i + 1}/{total})...")
        
        try:
            youtube.subscriptions().delete(id=sub_id).execute()
            print(f"Successfully unsubscribed from {channel_title}.")
            time.sleep(1)
            i += 1
        except HttpError as e:
            is_quota_error = False
            if e.resp.status == 403:
                try:
                    error_details = json.loads(e.content.decode('utf-8'))
                    if 'error' in error_details and 'errors' in error_details['error']:
                        for error in error_details['error']['errors']:
                            if error.get('reason') == 'quotaExceeded':
                                is_quota_error = True
                                break
                except (json.JSONDecodeError, KeyError):
                    pass
            
            if is_quota_error:
                wait_for_quota_reset()
                continue
            else:
                print(f"An HTTP error {e.resp.status} occurred while unsubscribing: {e.content}")
                i += 1
        except Exception as e:
            print(f"An unexpected error occurred: {e}")
            i += 1
            time.sleep(1)

def main():
    """Authenticates and unsubscribes from all YouTube subscriptions."""
    try:
        youtube = authenticate_youtube()
        subscriptions = list_subscriptions(youtube)
        
        if subscriptions:
            print(f"Found {len(subscriptions)} subscriptions.")
            unsubscribe_from_all(youtube, subscriptions)
            print("Finished unsubscribing from all channels.")
        else:
            print("Could not retrieve subscriptions or you have no subscriptions.")
            
    except Exception as e:
        print(f"An error occurred: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
