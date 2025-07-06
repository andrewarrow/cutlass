#!/usr/bin/env python3

import os
import sys
import pickle
import time
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

def list_subscriptions(youtube):
    """List all subscriptions for the authenticated user."""
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
            print(f"An HTTP error {e.resp.status} occurred: {e.content}")
            break
            
    return subscriptions

def unsubscribe_from_all(youtube, subscriptions):
    """Unsubscribe from all channels in the list."""
    total = len(subscriptions)
    for i, item in enumerate(subscriptions):
        sub_id = item['id']
        channel_title = item['snippet']['title']
        print(f"Unsubscribing from {channel_title} ({i+1}/{total})...")
        try:
            youtube.subscriptions().delete(id=sub_id).execute()
            print(f"Successfully unsubscribed from {channel_title}.")
        except HttpError as e:
            print(f"An HTTP error {e.resp.status} occurred while unsubscribing: {e.content}")
        
        # Adhere to rate limits by waiting a second between requests
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
